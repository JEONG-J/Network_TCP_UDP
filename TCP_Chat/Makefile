CC = gcc
CFLAGS = -Wall -Wextra -g

RUST_CLIENT_DIR = chatClient_rust
RUST_SERVER_DIR = chatServer_rust

all: chatClient chatServer build_rust_client build_rust_server

chatServer: chatServer.c
	$(CC) $(CFLAGS) -o chatServer chatServer.c

chatClient: chatClient.c
	$(CC) $(CFLAGS) -o chatClient chatClient.c

build_rust_client:
	cd $(RUST_CLIENT_DIR) && cargo build

build_rust_server:
	cd $(RUST_SERVER_DIR) && cargo build

run_rust_client:
	cd $(RUST_CLIENT_DIR) && cargo run

run_rust_server:
	cd $(RUST_SERVER_DIR) && cargo run

clean:
	rm -f chatServer chatClient
	cd $(RUST_CLIENT_DIR) && cargo clean
	cd $(RUST_SERVER_DIR) && cargo clean
