package factories

// UserSeed holds plaintext credentials for one demo user.
type UserSeed struct {
	Email     string
	Name      string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

// Users are 5 demo users representing common Indonesian employment profiles.
// All share Password123 for easy dev/demo access.
var Users = []UserSeed{
	{Email: "budi.santoso@gmail.com", Name: "Budi Santoso", Password: "Password123", FirstName: "Budi", LastName: "Santoso", Phone: "+6281234567890"},
	{Email: "siti.rahayu@gmail.com", Name: "Siti Rahayu", Password: "Password123", FirstName: "Siti", LastName: "Rahayu", Phone: "+6282345678901"},
	{Email: "ahmad.fauzi@gmail.com", Name: "Ahmad Fauzi", Password: "Password123", FirstName: "Ahmad", LastName: "Fauzi", Phone: "+6283456789012"},
	{Email: "dewi.permata@gmail.com", Name: "Dewi Permata", Password: "Password123", FirstName: "Dewi", LastName: "Permata", Phone: "+6284567890123"},
	{Email: "eko.prabowo@gmail.com", Name: "Eko Prabowo", Password: "Password123", FirstName: "Eko", LastName: "Prabowo", Phone: "+6285678901234"},
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

// NotificationPrefSeed holds notification preference settings per user.
type NotificationPrefSeed struct {
	BudgetAlerts  bool
	GoalReminders bool
	AnomalyAlerts bool
	WeeklySummary bool
	PushEnabled   bool
}

// NotificationPrefs index-aligned with Users.
var NotificationPrefs = []NotificationPrefSeed{
	// Budi — full notifications enabled
	{BudgetAlerts: true, GoalReminders: true, AnomalyAlerts: true, WeeklySummary: true, PushEnabled: true},
	// Siti — budget and goal alerts only
	{BudgetAlerts: true, GoalReminders: true, AnomalyAlerts: false, WeeklySummary: false, PushEnabled: false},
	// Ahmad — all enabled except push
	{BudgetAlerts: true, GoalReminders: true, AnomalyAlerts: true, WeeklySummary: true, PushEnabled: false},
	// Dewi — all enabled
	{BudgetAlerts: true, GoalReminders: true, AnomalyAlerts: true, WeeklySummary: false, PushEnabled: true},
	// Eko — minimal
	{BudgetAlerts: true, GoalReminders: true, AnomalyAlerts: false, WeeklySummary: false, PushEnabled: false},
}

// NotificationSeed holds data for one notification row.
type NotificationSeed struct {
	Type       string
	Title      string
	Message    string
	EntityType *string
	IsRead     bool
}

// Notifications index-aligned with Users.
var Notifications = [][]NotificationSeed{
	// Budi
	{
		{Type: "BUDGET_WARNING", Title: "Budget hampir habis: Makanan & Minuman", Message: "Anggaran Makanan & Minuman Anda telah mencapai 85%. Sisa anggaran: Rp 300.000.", EntityType: strPtr("budget"), IsRead: true},
		{Type: "BUDGET_EXCEEDED", Title: "Budget terlampaui: Transportasi", Message: "Anggaran Transportasi Anda bulan ini telah terlampaui sebesar 12%. Pertimbangkan untuk mengurangi pengeluaran transportasi.", EntityType: strPtr("budget"), IsRead: false},
		{Type: "GOAL_REMINDER", Title: "Pengingat: Dana Darurat", Message: "Anda telah mencapai 50% dari target Dana Darurat (Rp 15 juta dari Rp 30 juta). Deadline: 31 Desember 2026.", EntityType: strPtr("goal"), IsRead: false},
	},
	// Siti
	{
		{Type: "BUDGET_EXCEEDED", Title: "Budget terlampaui: Belanja", Message: "Anggaran Belanja Anda bulan ini telah terlampaui. Total pengeluaran belanja melebihi batas Rp 2.000.000.", EntityType: strPtr("budget"), IsRead: false},
		{Type: "GOAL_REMINDER", Title: "Pengingat: Modal Usaha Tambahan", Message: "Progress Modal Usaha Tambahan Anda: Rp 8 juta dari Rp 20 juta (40%). Deadline: 1 Januari 2027.", EntityType: strPtr("goal"), IsRead: true},
		{Type: "BUDGET_WARNING", Title: "Budget hampir habis: Tagihan", Message: "Anggaran Tagihan Anda telah mencapai 92%. Sisa anggaran: Rp 80.000.", EntityType: strPtr("budget"), IsRead: false},
	},
	// Ahmad
	{
		{Type: "GOAL_REMINDER", Title: "Pengingat: Biaya Umroh", Message: "Hampir mencapai target! Dana Umroh Anda sudah Rp 20 juta dari Rp 25 juta (80%). Deadline: 1 September 2026.", EntityType: strPtr("goal"), IsRead: false},
		{Type: "BUDGET_WARNING", Title: "Budget hampir habis: Utilitas", Message: "Anggaran Utilitas Anda telah mencapai 91%. Sisa anggaran: Rp 45.000.", EntityType: strPtr("budget"), IsRead: true},
		{Type: "GOAL_REMINDER", Title: "Pengingat: Dana Pendidikan Anak", Message: "Progress Dana Pendidikan Anak: Rp 5 juta dari Rp 50 juta (10%). Mulai tingkatkan kontribusi bulanan.", EntityType: strPtr("goal"), IsRead: false},
	},
	// Dewi
	{
		{Type: "BUDGET_EXCEEDED", Title: "Budget terlampaui: Hiburan & Rekreasi", Message: "Anggaran Hiburan & Rekreasi Anda bulan ini telah terlampaui. Kurangi pengeluaran hiburan bulan depan.", EntityType: strPtr("budget"), IsRead: false},
		{Type: "GOAL_REMINDER", Title: "Pengingat: Laptop MacBook Pro M3", Message: "Progress MacBook Pro M3: Rp 13 juta dari Rp 22 juta (59%). Deadline: 1 Agustus 2026 — segera capai target!", EntityType: strPtr("goal"), IsRead: false},
		{Type: "BUDGET_WARNING", Title: "Budget hampir habis: Belanja", Message: "Anggaran Belanja Anda telah mencapai 78%. Sisa anggaran: Rp 550.000.", EntityType: strPtr("budget"), IsRead: true},
	},
	// Eko
	{
		{Type: "BUDGET_WARNING", Title: "Budget hampir habis: Makanan & Minuman", Message: "Anggaran Makanan & Minuman Anda telah mencapai 82%. Sisa anggaran: Rp 144.000.", EntityType: strPtr("budget"), IsRead: false},
		{Type: "GOAL_REMINDER", Title: "Pengingat: Motor Honda Beat", Message: "Progress Motor Honda Beat: Rp 4 juta dari Rp 10 juta (40%). Deadline: 31 Desember 2026.", EntityType: strPtr("goal"), IsRead: true},
	},
}

// ActivityLogSeed holds data for one activity log row.
type ActivityLogSeed struct {
	Action      string
	EntityType  string
	Description string
}

// ActivityLogs index-aligned with Users.
var ActivityLogs = [][]ActivityLogSeed{
	// Budi
	{
		{Action: "CREATE", EntityType: "transaction", Description: "Created income transaction: Gaji 10000000"},
		{Action: "CREATE", EntityType: "budget", Description: "Created monthly budget for Makanan & Minuman"},
		{Action: "CREATE", EntityType: "goal", Description: "Created goal: Dana Darurat"},
		{Action: "UPDATE", EntityType: "goal", Description: "Updated contribution to Dana Darurat: Rp 15.000.000"},
		{Action: "CREATE", EntityType: "transaction", Description: "Created expense transaction: Transportasi 35000"},
		{Action: "UPDATE", EntityType: "budget", Description: "Updated budget limit for Hiburan & Rekreasi to Rp 1.000.000"},
	},
	// Siti
	{
		{Action: "CREATE", EntityType: "transaction", Description: "Created income transaction: Gaji 7500000"},
		{Action: "CREATE", EntityType: "budget", Description: "Created monthly budget for Belanja"},
		{Action: "CREATE", EntityType: "goal", Description: "Created goal: Modal Usaha Tambahan"},
		{Action: "UPDATE", EntityType: "goal", Description: "Updated contribution to Modal Usaha Tambahan: Rp 8.000.000"},
		{Action: "CREATE", EntityType: "transaction", Description: "Created income transaction: Freelance 2500000"},
	},
	// Ahmad
	{
		{Action: "CREATE", EntityType: "transaction", Description: "Created income transaction: Gaji 6000000"},
		{Action: "CREATE", EntityType: "goal", Description: "Created goal: Biaya Umroh"},
		{Action: "UPDATE", EntityType: "goal", Description: "Updated contribution to Biaya Umroh: Rp 20.000.000"},
		{Action: "CREATE", EntityType: "goal", Description: "Created goal: Dana Pendidikan Anak"},
		{Action: "UPDATE", EntityType: "goal", Description: "Completed goal: Renovasi Dapur"},
		{Action: "CREATE", EntityType: "budget", Description: "Created yearly budget for Pendidikan"},
	},
	// Dewi
	{
		{Action: "CREATE", EntityType: "transaction", Description: "Created income transaction: Gaji 8500000"},
		{Action: "CREATE", EntityType: "transaction", Description: "Created income transaction: Freelance 3000000"},
		{Action: "CREATE", EntityType: "goal", Description: "Created goal: Laptop MacBook Pro M3"},
		{Action: "UPDATE", EntityType: "goal", Description: "Updated contribution to Laptop MacBook Pro M3: Rp 13.000.000"},
		{Action: "CREATE", EntityType: "budget", Description: "Created yearly budget for Peralatan Kerja"},
		{Action: "DELETE", EntityType: "transaction", Description: "Deleted duplicate expense transaction"},
	},
	// Eko
	{
		{Action: "CREATE", EntityType: "transaction", Description: "Created income transaction: Gaji 2500000"},
		{Action: "CREATE", EntityType: "budget", Description: "Created monthly budget for Pendidikan"},
		{Action: "CREATE", EntityType: "goal", Description: "Created goal: Motor Honda Beat"},
		{Action: "UPDATE", EntityType: "goal", Description: "Updated contribution to Motor Honda Beat: Rp 4.000.000"},
		{Action: "CREATE", EntityType: "goal", Description: "Created goal: Laptop untuk Kuliah"},
	},
}

func intPtr(v int) *int    { return &v }
func strPtr(v string) *string { return &v }
