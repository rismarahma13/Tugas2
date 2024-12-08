 package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Struct untuk menyimpan item di menu
type ItemMenu struct {
	Nama  string
	Harga float64
}

// Struct untuk menangani pesanan
type Pesanan struct {
	Item   ItemMenu
	Jumlah int
}

// Fungsi goroutine untuk memproses pesanan
func prosesPesanan(id int, pesananCh <-chan Pesanan, wg *sync.WaitGroup) {
	defer wg.Done()
	for pesanan := range pesananCh {
		fmt.Printf("Goroutine %d sedang memproses: %s - Jumlah: %d\n", id, pesanan.Item.Nama, pesanan.Jumlah)
		time.Sleep(2 * time.Second) // Simulasi proses memakan waktu
	}
}

// Fungsi untuk validasi input harga dengan regexp
func cekHarga(input string) (float64, error) {
	formatHarga := regexp.MustCompile(^\d+(\.\d{1,2})?$)
	if !formatHarga.MatchString(input) {
		return 0, errors.New("format harga tidak valid")
	}
	harga, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, err
	}
	return harga, nil
}

// Fungsi untuk mengenkripsi detail pesanan (list item dan total harga)
func enkripsiDetailPesanan(pesananList []Pesanan, totalHarga float64) string {
	var rincian []string
	for _, pesanan := range pesananList {
		rincian = append(rincian, fmt.Sprintf("%s x%d (Rp%.2f)", pesanan.Item.Nama, pesanan.Jumlah, pesanan.Item.Harga*float64(pesanan.Jumlah)))
	}
	rincian = append(rincian, fmt.Sprintf("Total Biaya: Rp%.2f", totalHarga))
	data := strings.Join(rincian, "; ")
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return encoded
}

// Fungsi untuk mengambil input dari user
func ambilInput(pesan string) string {
	var input string
	fmt.Print(pesan)
	fmt.Scanln(&input)
	return input
}

// Fungsi untuk mengambil jumlah pesanan dari user
func ambilJumlahPesanan() (int, error) {
	jumlahStr := ambilInput("Masukkan jumlah pesanan: ")
	jumlah, err := strconv.Atoi(jumlahStr)
	if err != nil || jumlah <= 0 {
		return 0, errors.New("jumlah pesanan tidak valid")
	}
	return jumlah, nil
}

// Fungsi untuk menghitung kembalian
func hitungKembalian(total, bayar int) (int, error) {
	if bayar < total {
		return 0, errors.New("uang yang dibayarkan tidak mencukupi")
	}
	return bayar - total, nil
}

func main() {
	// Defer untuk memberikan pesan penutup
	defer fmt.Println("Program selesai dijalankan.")

	// Error handling menggunakan panic & recover
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi error:", r)
		}
	}()

	// Daftar menu makanan
	menuMakanan := map[string]ItemMenu{
		"nasgor": {Nama: "Nasi Goreng", Harga: 25000},
		"migor":  {Nama: "Mie Goreng", Harga: 22000},
		"ayamba": {Nama: "Ayam Bakar", Harga: 30000},
	}

	fmt.Println("Daftar Menu:")
	for _, item := range menuMakanan {
		fmt.Printf("- %s: Rp%.2f\n", item.Nama, item.Harga)
	}

	// Inisialisasi channel dan WaitGroup untuk goroutine
	pesananCh := make(chan Pesanan, 2)
	var wg sync.WaitGroup

	// Menjalankan worker goroutine
	wg.Add(1)
	go prosesPesanan(1, pesananCh, &wg)

	var totalHarga float64
	var daftarPesanan []Pesanan

	for {
		namaItem := ambilInput("Masukkan nama menu (ketik 'selesai' untuk berhenti): ")

		// Membersihkan input dari spasi dan mengubah menjadi huruf kecil
		namaItem = strings.ToLower(strings.TrimSpace(namaItem))

		if namaItem == "selesai" {
			break
		}

		item, ada := menuMakanan[namaItem]
		if !ada {
			fmt.Println("Item tidak ada dalam daftar menu. Silakan coba lagi.")
			continue
		}

		jumlah, err := ambilJumlahPesanan()
		if err != nil {
			fmt.Println(err)
			continue
		}

		pesanan := Pesanan{
			Item:   item,
			Jumlah: jumlah,
		}
		daftarPesanan = append(daftarPesanan, pesanan)
		totalHarga += item.Harga * float64(jumlah)

		pesananCh <- pesanan
	}

	close(pesananCh)
	wg.Wait()

	fmt.Println("\nDaftar Pesanan Anda:")
	for _, pesanan := range daftarPesanan {
		fmt.Printf("- %s\n", pesanan.Item.Nama)
	}
	fmt.Printf("Total Harga: Rp%.2f\n", totalHarga)

	// Mengenkripsi detail pesanan
	if len(daftarPesanan) > 0 {
		encodedDetail := enkripsiDetailPesanan(daftarPesanan, totalHarga)
		fmt.Println("Detail Pesanan (encoded base64):", encodedDetail)
	}

	// Meminta pembayaran dari user
	jumlahBayarStr := ambilInput("Masukkan jumlah uang yang dibayarkan: ")
	jumlahBayar, err := strconv.Atoi(jumlahBayarStr)
	if err != nil {
		panic("Input tidak valid!")
	}

	kembalian, err := hitungKembalian(int(totalHarga), jumlahBayar)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Pembayaran valid. Kembalian: Rp %d\n", kembalian)
}