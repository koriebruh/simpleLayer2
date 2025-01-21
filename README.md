
## Cara Menjalankan

### 1. **Menjalankan Server dan Client Menggunakan Makefile**

Proyek ini sudah dilengkapi dengan **Makefile** yang memudahkan Anda untuk menjalankan server dan client. Anda cukup menggunakan perintah `make` untuk menjalankan server atau client.

#### a. **Menjalankan Server**

Untuk menjalankan server, gunakan perintah berikut:

```bash
make run-server
```

Perintah ini akan **menjalankan server** yang ada di direktori `/server/main.go`.

#### b. **Menjalankan Client**

Untuk menjalankan client, gunakan perintah berikut:

```bash
make run-client
```

Perintah ini akan **menjalankan client** yang ada di **root proyek** (`main.go`).

#### c. **Menjalankan Server dan Client Secara Bersamaan**

Jika ingin menjalankan **server** dan **client** secara bersamaan, jalankan:

```bash
make run-all
```

Perintah ini akan menjalankan **server** dan **client** berturut-turut.

---
## Soon Improvement
### 1. **Mengurangi Dana dari Pengirim (Sender)**
- Saat ini, dana untuk transaksi batch dikurangi dari sistem (system) alih-alih dari pengirim (sender). Hal ini menyebabkan masalah jika dana pengirim tidak dikurangi saat transaksi diproses.

**Solusi:**
  Ubah logika transaksi sehingga saldo pengirim berkurang ketika transaksi dilakukan, bukan saldo sistem. Ini bisa dilakukan dengan memastikan bahwa transaksi yang ditandatangani dan dikirim mencerminkan pengurangan saldo dari pengirim, bukan sistem.
### 2. **Handling Error Lebih Baik**
- Perbaiki penanganan error dengan lebih detail, misalnya jika koneksi ke Ethereum gagal, tampilkan pesan yang lebih spesifik mengenai alasan kegagalan (misalnya masalah dengan koneksi Infura atau jaringan).
dan VALIDATE Fund Sender and User be

### 3. **Automasi Pengujian (Testing)**
- Tambahkan pengujian unit atau integrasi untuk memastikan bahwa **batch transactions** diproses dengan benar di server dan **client dapat mengirim transaksi dengan benar**.

### 4. **Peningkatan Scalability**
- Meningkatkan **skabilitas** server untuk menangani lebih banyak transaksi batch secara bersamaan, seperti dengan **concurrency** atau **queue-based processing**.

### 5. **Dokumentasi API gRPC**
- Menambahkan dokumentasi API **gRPC** yang lebih lengkap, sehingga orang lain dapat dengan mudah memahami cara berinteraksi dengan server Layer2 ini.

---