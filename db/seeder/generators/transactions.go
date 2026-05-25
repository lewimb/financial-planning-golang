package generators

import (
	"fmt"
	"math/rand"
	"time"
)

// TX is one generated transaction record.
type TX struct {
	Amount             int
	Category           string
	Type               string
	Date               time.Time
	Description        string
	IsRecurring        bool
	RecurrenceInterval string
}

// UserNames for logging, aligned with factories.Users.
var UserNames = []string{"Budi", "Siti", "Ahmad", "Dewi", "Eko"}

// baseSalaries in IDR per user index (aligned with factories.Users).
var baseSalaries = []int{
	10_000_000, // Budi — Software Engineer
	7_500_000,  // Siti — Wiraswasta
	6_000_000,  // Ahmad — PNS
	8_500_000,  // Dewi — Freelancer
	2_500_000,  // Eko — Mahasiswa
}

type txTemplate struct {
	category    string
	description string
	min, max    int
	isRecurring bool
	interval    string
}

var foodTemplates = []txTemplate{
	{category: "Makanan & Minuman", description: "Makan siang di warteg", min: 12_000, max: 35_000},
	{category: "Makanan & Minuman", description: "Makan malam di restoran", min: 45_000, max: 150_000},
	{category: "Makanan & Minuman", description: "Sarapan nasi uduk", min: 10_000, max: 22_000},
	{category: "Makanan & Minuman", description: "Kopi dan snack", min: 22_000, max: 65_000},
	{category: "Makanan & Minuman", description: "Belanja groceries mingguan", min: 120_000, max: 400_000},
	{category: "Makanan & Minuman", description: "Pesan GoFood/GrabFood", min: 35_000, max: 85_000},
	{category: "Makanan & Minuman", description: "Beli buah dan sayur", min: 30_000, max: 80_000},
	{category: "Makanan & Minuman", description: "Makan di food court mall", min: 40_000, max: 120_000},
}

var transportTemplates = []txTemplate{
	{category: "Transportasi", description: "Grab ke kantor", min: 15_000, max: 42_000},
	{category: "Transportasi", description: "Ojek online (Gojek)", min: 12_000, max: 35_000},
	{category: "Transportasi", description: "KRL Commuter Line", min: 4_000, max: 8_000},
	{category: "Transportasi", description: "Isi bensin motor", min: 50_000, max: 100_000},
	{category: "Transportasi", description: "Parkir kendaraan", min: 5_000, max: 15_000},
	{category: "Transportasi", description: "Bus Transjakarta", min: 3_500, max: 6_000},
	{category: "Transportasi", description: "Taksi online ke bandara", min: 80_000, max: 180_000},
}

var otherTemplates = []txTemplate{
	{category: "Hiburan & Rekreasi", description: "Nonton bioskop CGV", min: 55_000, max: 100_000},
	{category: "Hiburan & Rekreasi", description: "Langganan Spotify/Netflix", min: 54_000, max: 169_000, isRecurring: true, interval: "MONTHLY"},
	{category: "Hiburan & Rekreasi", description: "Karaoke bersama teman", min: 80_000, max: 200_000},
	{category: "Hiburan & Rekreasi", description: "Main game mobile (top-up)", min: 50_000, max: 200_000},
	{category: "Belanja", description: "Beli pakaian di Tokopedia", min: 100_000, max: 500_000},
	{category: "Belanja", description: "Belanja perlengkapan rumah", min: 150_000, max: 800_000},
	{category: "Belanja", description: "Beli skincare/toiletries", min: 100_000, max: 350_000},
	{category: "Belanja", description: "Beli peralatan dapur", min: 80_000, max: 400_000},
	{category: "Kesehatan", description: "Beli obat di apotek", min: 30_000, max: 150_000},
	{category: "Kesehatan", description: "Konsultasi dokter umum", min: 100_000, max: 350_000},
	{category: "Kesehatan", description: "Vitamin dan suplemen", min: 80_000, max: 250_000},
	{category: "Pendidikan", description: "Kursus online (Udemy/Coursera)", min: 80_000, max: 350_000},
	{category: "Pendidikan", description: "Beli buku teknis", min: 50_000, max: 250_000},
	{category: "Tagihan", description: "Tagihan telepon seluler", min: 100_000, max: 200_000, isRecurring: true, interval: "MONTHLY"},
	{category: "Tagihan", description: "Cicilan pinjaman lain", min: 200_000, max: 600_000, isRecurring: true, interval: "MONTHLY"},
}

var utilityTemplates = []txTemplate{
	{category: "Utilitas", description: "Tagihan listrik PLN", min: 200_000, max: 450_000, isRecurring: true, interval: "MONTHLY"},
	{category: "Utilitas", description: "Tagihan internet IndiHome", min: 200_000, max: 350_000, isRecurring: true, interval: "MONTHLY"},
	{category: "Utilitas", description: "Tagihan PDAM air", min: 80_000, max: 150_000, isRecurring: true, interval: "MONTHLY"},
	{category: "Utilitas", description: "Beli gas LPG 3kg", min: 20_000, max: 30_000, isRecurring: true, interval: "MONTHLY"},
}

// Generate produces deterministic transaction history for the given user index.
// Covers 7 months: Nov 2025 → May 2026 (current date 2026-05-19).
func Generate(userIndex int) []TX {
	rng := rand.New(rand.NewSource(int64(userIndex*99_991 + 12_345)))
	var result []TX

	base := baseSalaries[userIndex]
	now := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)

	for monthsBack := 6; monthsBack >= 0; monthsBack-- {
		ref := now.AddDate(0, -monthsBack, 0)
		monthStart := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, time.UTC)

		lastDay := daysInMonth(ref.Year(), ref.Month())
		if monthsBack == 0 && now.Day() < lastDay {
			lastDay = now.Day()
		}

		// INCOME — salary on 25th (or last available day)
		salaryDay := 25
		if lastDay < salaryDay {
			salaryDay = lastDay
		}
		variation := rng.Intn(300_000) - 150_000
		result = append(result, TX{
			Amount:             base + variation,
			Category:           "Gaji",
			Type:               "INCOME",
			Date:               date(monthStart.Year(), monthStart.Month(), salaryDay),
			Description:        fmt.Sprintf("Gaji bulan %s %d", monthStart.Month().String(), monthStart.Year()),
			IsRecurring:        true,
			RecurrenceInterval: "MONTHLY",
		})

		// INCOME — freelance for Siti (1) and Dewi (3)
		if (userIndex == 1 || userIndex == 3) && rng.Intn(3) < 2 {
			result = append(result, TX{
				Amount:      500_000 + rng.Intn(3_000_000),
				Category:    "Freelance",
				Type:        "INCOME",
				Date:        date(monthStart.Year(), monthStart.Month(), 1+rng.Intn(lastDay)),
				Description: "Pendapatan project freelance",
			})
		}

		// INCOME — occasional bonus (17% chance, not current month to avoid oddness)
		if monthsBack > 0 && rng.Intn(6) == 0 {
			result = append(result, TX{
				Amount:      1_000_000 + rng.Intn(4_000_000),
				Category:    "Bonus",
				Type:        "INCOME",
				Date:        date(monthStart.Year(), monthStart.Month(), 1+rng.Intn(lastDay)),
				Description: "Bonus kinerja tahunan",
			})
		}

		// EXPENSE — utility bills on 1st–5th
		for _, tpl := range utilityTemplates {
			// Eko (student) skips utilities (lives in kos, included in rent)
			if userIndex == 4 {
				continue
			}
			day := 1 + rng.Intn(5)
			if day > lastDay {
				day = lastDay
			}
			result = append(result, fromTemplate(rng, tpl, date(monthStart.Year(), monthStart.Month(), day)))
		}

		// EXPENSE — daily food and transport
		for day := 1; day <= lastDay; day++ {
			d := date(monthStart.Year(), monthStart.Month(), day)
			weekday := d.Weekday()

			// Food: 1–2 per day (skip 15% of days)
			if rng.Intn(20) > 2 {
				count := 1 + rng.Intn(2)
				for i := 0; i < count; i++ {
					tpl := foodTemplates[rng.Intn(len(foodTemplates))]
					result = append(result, fromTemplate(rng, tpl, d))
				}
			}

			// Transport: weekdays only (Eko uses commuter on weekdays)
			if weekday != time.Saturday && weekday != time.Sunday {
				count := 1 + rng.Intn(2)
				for i := 0; i < count; i++ {
					tpl := transportTemplates[rng.Intn(len(transportTemplates))]
					result = append(result, fromTemplate(rng, tpl, d))
				}
			}

			// Random other expense — 25% chance per day
			if rng.Intn(4) == 0 {
				tpl := otherTemplates[rng.Intn(len(otherTemplates))]
				result = append(result, fromTemplate(rng, tpl, d))
			}
		}
	}

	return result
}

func fromTemplate(rng *rand.Rand, tpl txTemplate, d time.Time) TX {
	return TX{
		Amount:             tpl.min + rng.Intn(tpl.max-tpl.min+1),
		Category:           tpl.category,
		Type:               "EXPENSE",
		Date:               d,
		Description:        tpl.description,
		IsRecurring:        tpl.isRecurring,
		RecurrenceInterval: tpl.interval,
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
