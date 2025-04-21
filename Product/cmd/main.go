package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	product "Taller2/Product/gen/go/api" // Ajusta esta importación según tu estructura de carpetas
	"Taller2/Product/internal/api"
	"Taller2/Product/internal/command"
	"Taller2/Product/internal/command/domain/entities"
	"Taller2/Product/internal/command/domain/events"
	kafkaComm "Taller2/Product/internal/command/infrastructure/kafka"
	"Taller2/Product/internal/command/infrastructure/persistence"
	"Taller2/Product/internal/query"
	kafkaQue "Taller2/Product/internal/query/infrastructure/kafka"
	persistenceQuery "Taller2/Product/internal/query/infrastructure/persistence"
	config "Taller2/Product/internal/shared"

	"github.com/google/uuid"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4321")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Soporte para preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func main() {
	// Cargar archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error cargando archivo .env:", err)
	}

	// Cargar configuración
	cfg := config.Load()

	log.Printf("Config loaded: %+v", cfg)

	// Inicializar Command y Query (como ya lo tienes configurado)
	commandRepo, err := persistence.NewSQLiteProductRepository(cfg.SQLite.CommandDSN)
	if err != nil {
		log.Fatal("Failed to init Command DB:", err)
	}

	// Crear un producto de ejemplo
	productID := uuid.New()
	exampleProduct := &entities.Product{
		ID:          productID, // Asume un ID único
		Name:        "Producto Ejemplo",
		Description: "Este es un producto de ejemplo",
		Price:       100.0,
		Stock:       50,
	}

	// Agregar el producto a la base de datos de Command
	err = commandRepo.Save(context.Background(), exampleProduct)
	if err != nil {
		log.Fatalf("Error agregando producto a la base de datos de Command: %v", err)
	}
	log.Println("Producto de ejemplo agregado exitosamente en Command DB")

	// Inicializar productor Kafka
	kafkaProducer := kafkaComm.NewProducer(cfg.Kafka.Brokers)

	// Crear el evento de producto creado
	event := events.NewProductCreatedEvent(exampleProduct)

	// Publicar el evento en Kafka
	err = kafkaProducer.PublishEvent(context.Background(), event)
	if err != nil {
		log.Printf("Error publicando evento a Kafka: %v", err)
	} else {
		log.Printf("Evento 'product_created' enviado a Kafka para el producto: %+v", event)
	}

	// Crear la aplicación de Command
	commandApp := command.NewApplication(commandRepo, kafkaProducer)

	// Inicializar repositorio de Query
	queryRepo, err := persistenceQuery.NewSQLiteProductReadRepository(cfg.SQLite.QueryDSN)
	if err != nil {
		log.Fatal("Failed to init Query DB:", err)
	}

	// Inicializar consumidor Kafka para Query
	kafkaConsumer := kafkaQue.NewConsumer(cfg.Kafka.Brokers, "product_events", queryRepo)

	// Iniciar consumidor de Kafka para Query (debe estar escuchando para los eventos)
	go func() {
		if err := kafkaConsumer.Start(context.Background()); err != nil {
			log.Fatalf("Error al iniciar consumidor de Kafka para Query: %v", err)
		}
	}()

	// Crear la aplicación de Query
	queryApp := query.NewApplication(queryRepo)

	// Iniciar servidor GRPC
	grpcServer := grpc.NewServer()
	product.RegisterProductServiceServer(
		grpcServer,
		api.NewProductServer(commandApp, queryApp),
	)

	// Iniciar servidor gRPC en paralelo
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Servicio Producto GRPC escuchando en %s", cfg.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// Iniciar gRPC-Gateway (HTTP REST)
	go func() {
		ctx := context.Background()
		mux := runtime.NewServeMux()

		// Conectar al servidor gRPC (especifica la dirección gRPC)
		opts := []grpc.DialOption{grpc.WithInsecure()}
		err := product.RegisterProductServiceHandlerFromEndpoint(
			ctx,
			mux,
			"localhost:"+cfg.GRPC.Port, // Conéctate al servidor gRPC
			opts,
		)
		if err != nil {
			log.Fatalf("Error iniciando gRPC-Gateway: %v", err)
		}

		corsHandler := allowCORS(mux)

		// Servir el tráfico HTTP en el puerto configurado
		log.Printf("gRPC-Gateway REST escuchando en %s", cfg.HTTP.Port)
		if err := http.ListenAndServe(":"+cfg.HTTP.Port, corsHandler); err != nil {
			log.Fatalf("Fallo servidor HTTP: %v", err)
		}
	}()

	// Mantener el servidor activo
	select {}
}
