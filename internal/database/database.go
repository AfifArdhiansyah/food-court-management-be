package database

import (
	"fmt"
	"log"

	"foodcourt-backend/internal/config"
	"foodcourt-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func New(cfg *config.Config) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	return &Database{DB: db}, nil
}

func (d *Database) Migrate() error {
	log.Println("Running database migrations...")

	err := d.DB.AutoMigrate(
		&models.User{},
		&models.Kios{},
		&models.Menu{},
		&models.Order{},
		&models.OrderItem{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (d *Database) Seed() error {
	log.Println("Running database seeder...")

	// Check if data already exists
	var userCount int64
	d.DB.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Create default cashier user
	cashierUser := &models.User{
		Username: "cashier",
		Email:    "cashier@foodcourt.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FullName: "Kasir Utama",
		Role:     models.RoleCashier,
		IsActive: true,
	}

	if err := d.DB.Create(cashierUser).Error; err != nil {
		return fmt.Errorf("failed to create cashier user: %w", err)
	}

	// Create sample kios
	kios1 := &models.Kios{
		Name:        "Warung Nasi Padang",
		Description: "Masakan Padang autentik dengan cita rasa tradisional",
		Location:    "Blok A-1",
		IsActive:    true,
	}

	kios2 := &models.Kios{
		Name:        "Kedai Mie Ayam",
		Description: "Mie ayam dan bakso dengan kuah yang gurih",
		Location:    "Blok A-2",
		IsActive:    true,
	}

	if err := d.DB.Create([]*models.Kios{kios1, kios2}).Error; err != nil {
		return fmt.Errorf("failed to create sample kios: %w", err)
	}

	// Create kios users
	kiosUser1 := &models.User{
		Username: "padang_user",
		Email:    "padang@foodcourt.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FullName: "Pelayan Warung Padang",
		Role:     models.RoleKios,
		KiosID:   &kios1.ID,
		IsActive: true,
	}

	kiosUser2 := &models.User{
		Username: "mieayam_user",
		Email:    "mieayam@foodcourt.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FullName: "Pelayan Kedai Mie Ayam",
		Role:     models.RoleKios,
		KiosID:   &kios2.ID,
		IsActive: true,
	}

	if err := d.DB.Create([]*models.User{kiosUser1, kiosUser2}).Error; err != nil {
		return fmt.Errorf("failed to create kios users: %w", err)
	}

	// Create sample menus
	menus := []*models.Menu{
		// Warung Nasi Padang
		{KiosID: kios1.ID, Name: "Nasi Rendang", Description: "Nasi putih dengan rendang daging sapi", Price: 25000, Category: models.CategoryFood, IsAvailable: true},
		{KiosID: kios1.ID, Name: "Nasi Ayam Pop", Description: "Nasi putih dengan ayam pop khas Padang", Price: 22000, Category: models.CategoryFood, IsAvailable: true},
		{KiosID: kios1.ID, Name: "Gulai Kambing", Description: "Gulai kambing dengan bumbu rempah", Price: 30000, Category: models.CategoryFood, IsAvailable: true},
		{KiosID: kios1.ID, Name: "Es Teh Manis", Description: "Es teh manis segar", Price: 5000, Category: models.CategoryDrink, IsAvailable: true},
		{KiosID: kios1.ID, Name: "Es Jeruk", Description: "Es jeruk peras segar", Price: 8000, Category: models.CategoryDrink, IsAvailable: true},

		// Kedai Mie Ayam
		{KiosID: kios2.ID, Name: "Mie Ayam Bakso", Description: "Mie ayam dengan bakso sapi", Price: 15000, Category: models.CategoryFood, IsAvailable: true},
		{KiosID: kios2.ID, Name: "Mie Ayam Ceker", Description: "Mie ayam dengan ceker ayam", Price: 18000, Category: models.CategoryFood, IsAvailable: true},
		{KiosID: kios2.ID, Name: "Bakso Urat", Description: "Bakso urat dengan kuah kaldu", Price: 20000, Category: models.CategoryFood, IsAvailable: true},
		{KiosID: kios2.ID, Name: "Es Teh Tawar", Description: "Es teh tawar", Price: 3000, Category: models.CategoryDrink, IsAvailable: true},
		{KiosID: kios2.ID, Name: "Jus Jeruk", Description: "Jus jeruk segar", Price: 10000, Category: models.CategoryDrink, IsAvailable: true},
	}

	if err := d.DB.Create(menus).Error; err != nil {
		return fmt.Errorf("failed to create sample menus: %w", err)
	}

	log.Println("Database seeded successfully")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
