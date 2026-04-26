package budget

import (
	"time"

	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/repository"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// Seed populates the database with default budget categories
func Seed(db *gorm.DB, userID uint) error {
	categoryRepo := repository.NewCategoryRepository(db)

	// Default income categories based on the spreadsheet
	incomeCategories := []domain.Category{
		{UserID: userID, Name: "Salário", Type: domain.CategoryTypeIncome, Icon: "HandCoins", Color: "#4CAF50"},
		{UserID: userID, Name: "Freelance", Type: domain.CategoryTypeIncome, Icon: "Briefcase", Color: "#2196F3"},
		{UserID: userID, Name: "Aluguel de Imóvel Online", Type: domain.CategoryTypeIncome, Icon: "Home", Color: "#009688"},
		{UserID: userID, Name: "Investimentos", Type: domain.CategoryTypeIncome, Icon: "TrendingUp", Color: "#FF9800"},
		{UserID: userID, Name: "Premiações", Type: domain.CategoryTypeIncome, Icon: "Gift", Color: "#FFC107"},
		{UserID: userID, Name: "Outras Fontes", Type: domain.CategoryTypeIncome, Icon: "Banknote", Color: "#8BC34A"},
	}

	// Default expense categories based on the spreadsheet
	expenseCategories := []domain.Category{
		// Moradia
		{UserID: userID, Name: "Aluguel", Type: domain.CategoryTypeExpense, Icon: "Home", Color: "#F44336"},
		{UserID: userID, Name: "Condomínio", Type: domain.CategoryTypeExpense, Icon: "Building2", Color: "#E91E63"},
		{UserID: userID, Name: "Energia", Type: domain.CategoryTypeExpense, Icon: "Zap", Color: "#9C27B0"},
		{UserID: userID, Name: "Água", Type: domain.CategoryTypeExpense, Icon: "Droplet", Color: "#673AB7"},
		{UserID: userID, Name: "Internet", Type: domain.CategoryTypeExpense, Icon: "Wifi", Color: "#3F51B5"},
		{UserID: userID, Name: "Gás", Type: domain.CategoryTypeExpense, Icon: "Flame", Color: "#FF5722"},
		{UserID: userID, Name: "IPTU", Type: domain.CategoryTypeExpense, Icon: "Receipt", Color: "#795548"},
		{UserID: userID, Name: "Manutenção", Type: domain.CategoryTypeExpense, Icon: "Wrench", Color: "#607D8B"},

		// Food
		{UserID: userID, Name: "Mercado", Type: domain.CategoryTypeExpense, Icon: "ShoppingCart", Color: "#4CAF50"},
		{UserID: userID, Name: "Refeições Fora", Type: domain.CategoryTypeExpense, Icon: "UtensilsCrossed", Color: "#8BC34A"},
		{UserID: userID, Name: "Lanches/Cafés", Type: domain.CategoryTypeExpense, Icon: "Coffee", Color: "#CDDC39"},
		{UserID: userID, Name: "Delivery", Type: domain.CategoryTypeExpense, Icon: "ShoppingBag", Color: "#FFEB3B"},

		// Transporte
		{UserID: userID, Name: "Combustível", Type: domain.CategoryTypeExpense, Icon: "Fuel", Color: "#FF9800"},
		{UserID: userID, Name: "Uber", Type: domain.CategoryTypeExpense, Icon: "Car", Color: "#FF5722"},
		{UserID: userID, Name: "Transporte Público", Type: domain.CategoryTypeExpense, Icon: "Bus", Color: "#F44336"},
		{UserID: userID, Name: "Manutenção Veículo", Type: domain.CategoryTypeExpense, Icon: "Wrench", Color: "#E91E63"},
		{UserID: userID, Name: "Seguro Auto", Type: domain.CategoryTypeExpense, Icon: "Car", Color: "#9C27B0"},
		{UserID: userID, Name: "IPVA", Type: domain.CategoryTypeExpense, Icon: "FileText", Color: "#673AB7"},
		{UserID: userID, Name: "Estacionamento/Pedágios", Type: domain.CategoryTypeExpense, Icon: "Car", Color: "#3F51B5"},

		// Health
		{UserID: userID, Name: "Plano de Saúde", Type: domain.CategoryTypeExpense, Icon: "HeartPulse", Color: "#2196F3"},
		{UserID: userID, Name: "Medicamentos", Type: domain.CategoryTypeExpense, Icon: "Pill", Color: "#03A9F4"},
		{UserID: userID, Name: "Consultas/Exames", Type: domain.CategoryTypeExpense, Icon: "Stethoscope", Color: "#00BCD4"},
		{UserID: userID, Name: "Academia", Type: domain.CategoryTypeExpense, Icon: "Dumbbell", Color: "#009688"},
		{UserID: userID, Name: "Terapia/Psicólogo", Type: domain.CategoryTypeExpense, Icon: "HeartPulse", Color: "#4CAF50"},

		// Education
		{UserID: userID, Name: "Cursos", Type: domain.CategoryTypeExpense, Icon: "GraduationCap", Color: "#8BC34A"},
		{UserID: userID, Name: "Livros/Material", Type: domain.CategoryTypeExpense, Icon: "BookOpen", Color: "#CDDC39"},
		{UserID: userID, Name: "Assinaturas Educacionais", Type: domain.CategoryTypeExpense, Icon: "GraduationCap", Color: "#FFEB3B"},
		{UserID: userID, Name: "Mensalidades/Escola", Type: domain.CategoryTypeExpense, Icon: "GraduationCap", Color: "#FFC107"},

		// Lazer
		{UserID: userID, Name: "Streaming", Type: domain.CategoryTypeExpense, Icon: "Film", Color: "#FFC107"},
		{UserID: userID, Name: "Viagens/Passeios", Type: domain.CategoryTypeExpense, Icon: "Plane", Color: "#FF9800"},
		{UserID: userID, Name: "Hobbies", Type: domain.CategoryTypeExpense, Icon: "Gamepad2", Color: "#FF5722"},
		{UserID: userID, Name: "Restaurantes", Type: domain.CategoryTypeExpense, Icon: "UtensilsCrossed", Color: "#F44336"},
		{UserID: userID, Name: "Cinema/Teatro", Type: domain.CategoryTypeExpense, Icon: "Film", Color: "#E91E63"},

		// Pessoal
		{UserID: userID, Name: "Roupas", Type: domain.CategoryTypeExpense, Icon: "Shirt", Color: "#9C27B0"},
		{UserID: userID, Name: "Beleza/Estética", Type: domain.CategoryTypeExpense, Icon: "Camera", Color: "#673AB7"},
		{UserID: userID, Name: "Presentes", Type: domain.CategoryTypeExpense, Icon: "Gift", Color: "#3F51B5"},
		{UserID: userID, Name: "Pets", Type: domain.CategoryTypeExpense, Icon: "PawPrint", Color: "#2196F3"},

		// Genérico (fallback para importação)
		{UserID: userID, Name: "Outros", Type: domain.CategoryTypeExpense, Icon: "Wallet", Color: "#6366F1"},
	}

	// Check if categories already exist for this user
	existingCategories, err := categoryRepo.GetByUserID(db.Statement.Context, userID)
	if err != nil {
		return err
	}

	// If user already has categories, skip seeding
	if len(existingCategories) > 0 {
		return nil
	}

	// Create income categories
	for _, category := range incomeCategories {
		category.CreatedAt = time.Now()
		category.UpdatedAt = time.Now()
		if _, err := categoryRepo.Create(db.Statement.Context, category); err != nil {
			return err
		}
	}

	// Create expense categories
	for _, category := range expenseCategories {
		category.CreatedAt = time.Now()
		category.UpdatedAt = time.Now()
		if _, err := categoryRepo.Create(db.Statement.Context, category); err != nil {
			return err
		}
	}

	return nil
}
