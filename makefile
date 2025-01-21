# Nama file Go untuk server dan client
#SERVER = server/main.go  got bug if do this
SERVER = main.go
CLIENT = main.go

# Go binary
GO = go

# Target untuk menjalankan server
.PHONY: run-server
run-server:
	@echo "Menjalankan server..."
	cd server && $(GO) run $(SERVER)

# Target untuk menjalankan client
.PHONY: run-client
run-client:
	@echo "Menjalankan client..."
	$(GO) run $(CLIENT)

# Target untuk menjalankan server dan client secara bersamaan (opsional)
.PHONY: run-all
run-all: run-server run-client

# Target untuk membersihkan build (tidak ada build yang dihasilkan, tapi bisa digunakan untuk membersihkan file lain)
.PHONY: clean
clean:
	@echo "Membersihkan..."
	rm -rf *.tmp
