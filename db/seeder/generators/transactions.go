package generators

import (
	"fmt"
	"math/rand"
	"time"
)

// TX is one generated transaction record.
type TX struct {
	Amount      int
	Category    string
	Type        string
	Date        time.Time
	Description string
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
}

var foodTemplates = []txTemplate{
	{"Makanan & Minuman", "Makan siang di warteg", 12_000, 35_000},
	{"Makanan & Minuman", "Makan malam di restoran", 45_000, 150_000},
	{"Makanan & Minuman", "Sarapan nasi uduk", 10_000, 22_000},
	{"Makanan & Minuman", "Kopi dan snack", 22_000, 65_000},
	{"Makanan & Minuman", "Belanja groceries mingguan", 120_000, 400_000},
	{"Makanan & Minuman", "Pesan GoFood/GrabFood", 35_000, 85_000},
	{"Makanan & Minuman", "Beli buah dan sayur", 30_000, 80_000},
	{"Makanan & Minuman", "Makan di food court mall", 40_000, 120_000},
}

var transportTemplates = []txTemplate{
	{"Transportasi", "Grab ke kantor", 15_000, 42_000},
	{"Transportasi", "Ojek online (Gojek)", 12_000, 35_000},
	{"Transportasi", "KRL Commuter Line", 4_000, 8_000},
	{"Transportasi", "Isi bensin motor", 50_000, 100_000},
	{"Transportasi", "Parkir kendaraan", 5_000, 15_000},
	{"Transportasi", "Bus Transjakarta", 3_500, 6_000},
	{"Transportasi", "Taksi online ke bandara", 80_000, 180_000},
}

var otherTemplates = []txTemplate{
	{"Hiburan & Rekreasi", "Nonton bioskop CGV", 55_000, 100_000},
	{"Hiburan & Rekreasi", "Langganan Spotify/Netflix", 54_000, 169_000},
	{"Hiburan & Rekreasi", "Karaoke bersama teman", 80_000, 200_000},
	{"Hiburan & Rekreasi", "Main game mobile (top-up)", 50_000, 200_000},
	{"Belanja", "Beli pakaian di Tokopedia", 100_000, 500_000},
	{"Belanja", "Belanja perlengkapan rumah", 150_000, 800_000},
	{"Belanja", "Beli skincare/toiletries", 100_000, 350_000},
	{"Belanja", "Beli peralatan dapur", 80_000, 400_000},
	{"Kesehatan", "Beli obat di apotek", 30_000, 150_000},
	{"Kesehatan", "Konsultasi dokter umum", 100_000, 350_000},
	{"Kesehatan", "Vitamin dan suplemen", 80_000, 250_000},
	{"Pendidikan", "Kursus online (Udemy/Coursera)", 80_000, 350_000},
	{"Pendidikan", "Beli buku teknis", 50_000, 250_000},
	{"Tagihan", "Tagihan telepon seluler", 100_000, 200_000},
	{"Tagihan", "Cicilan pinjaman lain", 200_000, 600_000},
}

var utilityTemplates = []txTemplate{
	{"Utilitas", "Tagihan listrik PLN", 200_000, 450_000},
	{"Utilitas", "Tagihan internet IndiHome", 200_000, 350_000},
	{"Utilitas", "Tagihan PDAM air", 80_000, 150_000},
	{"Utilitas", "Beli gas LPG 3kg", 20_000, 30_000},
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
			Amount:      base + variation,
			Category:    "Gaji",
			Type:        "INCOME",
			Date:        date(monthStart.Year(), monthStart.Month(), salaryDay),
			Description: fmt.Sprintf("Gaji bulan %s %d", monthStart.Month().String(), monthStart.Year()),
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
		Amount:      tpl.min + rng.Intn(tpl.max-tpl.min+1),
		Category:    tpl.category,
		Type:        "EXPENSE",
		Date:        d,
		Description: tpl.description,
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
