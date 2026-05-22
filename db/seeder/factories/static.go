package factories

// UserSeed holds plaintext credentials for one demo user.
type UserSeed struct {
	Email    string
	Name     string
	Password string
}

// Users are 5 demo users representing common Indonesian employment profiles.
// All share Password123 for easy dev/demo access.
var Users = []UserSeed{
	{Email: "budi.santoso@gmail.com", Name: "Budi Santoso", Password: "Password123"},
	{Email: "siti.rahayu@gmail.com", Name: "Siti Rahayu", Password: "Password123"},
	{Email: "ahmad.fauzi@gmail.com", Name: "Ahmad Fauzi", Password: "Password123"},
	{Email: "dewi.permata@gmail.com", Name: "Dewi Permata", Password: "Password123"},
	{Email: "eko.prabowo@gmail.com", Name: "Eko Prabowo", Password: "Password123"},
}

// ProfileSeed maps to UpsertFinancialProfileRequest fields.
type ProfileSeed struct {
	MonthlyIncome    float64
	FixedExpenses    float64
	CurrentSavings   float64
	Debt             float64
	EmploymentStatus string
	SpendingHabit    string
	RiskLevel        string
	Goals            []string
}

// Profiles index-aligned with Users.
var Profiles = []ProfileSeed{
	{
		// Budi — Software Engineer swasta
		MonthlyIncome:    10_000_000,
		FixedExpenses:    4_000_000,
		CurrentSavings:   25_000_000,
		Debt:             15_000_000,
		EmploymentStatus: "EMPLOYED",
		SpendingHabit:    "MODERATE",
		RiskLevel:        "MEDIUM",
		Goals:            []string{"EMERGENCY_FUND", "PROPERTY", "RETIREMENT"},
	},
	{
		// Siti — Pemilik toko online (wiraswasta)
		MonthlyIncome:    7_500_000,
		FixedExpenses:    3_500_000,
		CurrentSavings:   12_000_000,
		Debt:             8_000_000,
		EmploymentStatus: "SELF_EMPLOYED",
		SpendingHabit:    "MODERATE",
		RiskLevel:        "HIGH",
		Goals:            []string{"INVESTMENT", "VEHICLE", "EMERGENCY_FUND"},
	},
	{
		// Ahmad — PNS golongan III
		MonthlyIncome:    6_000_000,
		FixedExpenses:    2_500_000,
		CurrentSavings:   30_000_000,
		Debt:             0,
		EmploymentStatus: "CIVIL_SERVANT",
		SpendingHabit:    "FRUGAL",
		RiskLevel:        "LOW",
		Goals:            []string{"EDUCATION", "RETIREMENT"},
	},
	{
		// Dewi — UI/UX freelancer
		MonthlyIncome:    8_500_000,
		FixedExpenses:    3_000_000,
		CurrentSavings:   18_000_000,
		Debt:             5_000_000,
		EmploymentStatus: "FREELANCER",
		SpendingHabit:    "LIBERAL",
		RiskLevel:        "HIGH",
		Goals:            []string{"VACATION", "PROPERTY", "INVESTMENT"},
	},
	{
		// Eko — Mahasiswa part-time
		MonthlyIncome:    2_500_000,
		FixedExpenses:    1_500_000,
		CurrentSavings:   3_000_000,
		Debt:             2_000_000,
		EmploymentStatus: "STUDENT",
		SpendingHabit:    "FRUGAL",
		RiskLevel:        "LOW",
		Goals:            []string{"EDUCATION", "VEHICLE"},
	},
}

// BudgetSeed holds data for one budget row.
type BudgetSeed struct {
	Category       string
	Period         string
	Month          *int // nil for YEARLY
	Year           int
	LimitAmount    int
	AlertThreshold int
}

// Budgets index-aligned with Users — current month (May 2026) + yearly.
var Budgets = [][]BudgetSeed{
	// Budi
	{
		{Category: "Makanan & Minuman", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 2_000_000, AlertThreshold: 80},
		{Category: "Transportasi", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 800_000, AlertThreshold: 80},
		{Category: "Hiburan & Rekreasi", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 1_000_000, AlertThreshold: 75},
		{Category: "Belanja", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 1_500_000, AlertThreshold: 80},
		{Category: "Utilitas", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 600_000, AlertThreshold: 90},
		{Category: "Kesehatan", Period: "YEARLY", Month: nil, Year: 2026, LimitAmount: 5_000_000, AlertThreshold: 80},
	},
	// Siti
	{
		{Category: "Makanan & Minuman", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 1_500_000, AlertThreshold: 80},
		{Category: "Transportasi", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 600_000, AlertThreshold: 80},
		{Category: "Belanja", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 2_000_000, AlertThreshold: 85},
		{Category: "Tagihan", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 1_000_000, AlertThreshold: 90},
		{Category: "Investasi", Period: "YEARLY", Month: nil, Year: 2026, LimitAmount: 20_000_000, AlertThreshold: 50},
	},
	// Ahmad
	{
		{Category: "Makanan & Minuman", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 1_000_000, AlertThreshold: 80},
		{Category: "Transportasi", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 400_000, AlertThreshold: 80},
		{Category: "Utilitas", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 500_000, AlertThreshold: 90},
		{Category: "Pendidikan", Period: "YEARLY", Month: nil, Year: 2026, LimitAmount: 3_000_000, AlertThreshold: 75},
	},
	// Dewi
	{
		{Category: "Makanan & Minuman", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 1_800_000, AlertThreshold: 80},
		{Category: "Transportasi", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 700_000, AlertThreshold: 80},
		{Category: "Hiburan & Rekreasi", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 1_500_000, AlertThreshold: 75},
		{Category: "Belanja", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 2_500_000, AlertThreshold: 80},
		{Category: "Peralatan Kerja", Period: "YEARLY", Month: nil, Year: 2026, LimitAmount: 10_000_000, AlertThreshold: 80},
	},
	// Eko
	{
		{Category: "Makanan & Minuman", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 800_000, AlertThreshold: 80},
		{Category: "Transportasi", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 300_000, AlertThreshold: 80},
		{Category: "Pendidikan", Period: "MONTHLY", Month: intPtr(5), Year: 2026, LimitAmount: 500_000, AlertThreshold: 80},
	},
}

// GoalSeed holds data for one financial goal.
type GoalSeed struct {
	Name          string
	TargetAmount  int
	CurrentAmount int
	Status        string
	Deadline      string // YYYY-MM-DD
	Description   string
}

// Goals index-aligned with Users.
var Goals = [][]GoalSeed{
	// Budi
	{
		{Name: "Dana Darurat", TargetAmount: 30_000_000, CurrentAmount: 15_000_000, Status: "ONGOING", Deadline: "2026-12-31", Description: "Target 3x pengeluaran bulanan sebagai dana darurat keluarga"},
		{Name: "DP Rumah", TargetAmount: 100_000_000, CurrentAmount: 10_000_000, Status: "ONGOING", Deadline: "2028-06-30", Description: "Uang muka pembelian rumah di kawasan Bekasi Timur"},
		{Name: "Liburan ke Jepang", TargetAmount: 15_000_000, CurrentAmount: 7_500_000, Status: "ONGOING", Deadline: "2026-10-01", Description: "Liburan keluarga ke Tokyo dan Osaka selama 10 hari"},
	},
	// Siti
	{
		{Name: "Modal Usaha Tambahan", TargetAmount: 20_000_000, CurrentAmount: 8_000_000, Status: "ONGOING", Deadline: "2027-01-01", Description: "Tambah modal untuk ekspansi toko online ke platform baru"},
		{Name: "Beli Mobil Honda Brio", TargetAmount: 200_000_000, CurrentAmount: 20_000_000, Status: "ONGOING", Deadline: "2029-12-31", Description: "DP dan kredit mobil pertama untuk kebutuhan operasional usaha"},
	},
	// Ahmad
	{
		{Name: "Biaya Umroh", TargetAmount: 25_000_000, CurrentAmount: 20_000_000, Status: "ONGOING", Deadline: "2026-09-01", Description: "Biaya perjalanan umroh bersama istri"},
		{Name: "Dana Pendidikan Anak", TargetAmount: 50_000_000, CurrentAmount: 5_000_000, Status: "ONGOING", Deadline: "2030-07-01", Description: "Tabungan pendidikan anak untuk masuk perguruan tinggi"},
		{Name: "Renovasi Dapur", TargetAmount: 8_000_000, CurrentAmount: 8_000_000, Status: "COMPLETED", Deadline: "2025-12-31", Description: "Renovasi dapur dan ruang makan"},
	},
	// Dewi
	{
		{Name: "Laptop MacBook Pro M3", TargetAmount: 22_000_000, CurrentAmount: 13_000_000, Status: "ONGOING", Deadline: "2026-08-01", Description: "Upgrade laptop untuk pekerjaan desain UI/UX yang lebih profesional"},
		{Name: "Liburan Backpacker Eropa", TargetAmount: 30_000_000, CurrentAmount: 5_000_000, Status: "ONGOING", Deadline: "2027-06-01", Description: "Solo trip backpacker ke 5 negara Eropa selama 3 minggu"},
	},
	// Eko
	{
		{Name: "Motor Honda Beat", TargetAmount: 10_000_000, CurrentAmount: 4_000_000, Status: "ONGOING", Deadline: "2026-12-31", Description: "Motor bekas untuk transportasi kuliah sehari-hari"},
		{Name: "Laptop untuk Kuliah", TargetAmount: 8_000_000, CurrentAmount: 3_500_000, Status: "ONGOING", Deadline: "2026-09-01", Description: "Laptop baru untuk mendukung tugas kuliah dan skripsi"},
	},
}

// AiLogSeed is one sample AI chat entry (applied to all users).
type AiLogSeed struct {
	Question string
	Response string
}

var AiLogs = []AiLogSeed{
	{
		Question: "Berapa total pengeluaran saya bulan ini?",
		Response: "Berdasarkan data transaksi Anda bulan ini, total pengeluaran Anda sekitar Rp 3,2 juta. Kategori terbesar adalah Makanan & Minuman (Rp 1,1 juta) diikuti Transportasi (Rp 650 ribu). Pengeluaran Anda masih dalam batas wajar sesuai budget yang Anda tetapkan.",
	},
	{
		Question: "Apakah saya melebihi budget untuk makanan bulan ini?",
		Response: "Budget makanan Anda bulan ini Rp 1,5 juta, dan sudah terpakai sekitar 73% (Rp 1,1 juta). Anda masih punya sisa Rp 400 ribu. Dengan sisa hari bulan ini, batasi pengeluaran makan di bawah Rp 35 ribu/hari agar tetap dalam budget.",
	},
	{
		Question: "Kapan saya bisa mencapai target tabungan untuk dana darurat?",
		Response: "Berdasarkan rata-rata tabungan bulanan Anda saat ini, target dana darurat bisa tercapai dalam 6–8 bulan. Untuk mempercepat, pertimbangkan menambah penghasilan pasif atau mengurangi pengeluaran hiburan yang tidak esensial.",
	},
}

func intPtr(v int) *int { return &v }
