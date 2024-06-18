# Java configuration
JAVAC = javac
JAVA = java
CLIENT = SplitFileClient
SERVER = SplitFileServer

CLIENT_SRC = $(CLIENT).java
SERVER_SRC = $(SERVER).java
JAVA_BIN_DIR = bin

RUST_CLIENT_DIR = SplitFileClient_rust
RUST_SERVER_DIR = SplitFileServer_rust

# Rust configuration
CARGO = cargo
RUST_CLIENT = split_file_client
RUST_SERVER = split_file_server

all: java-client java-server build_rust_client build_rust_server

# Java build rules
java-client: $(CLIENT_SRC)
	$(JAVAC) $(CLIENT_SRC)

java-server: $(SERVER_SRC)
	$(JAVAC) $(SERVER_SRC)


# Rust build rules
build_rust_client:
	cd $(RUST_CLIENT_DIR) && $(CARGO) build

build_rust_server:
	cd $(RUST_SERVER_DIR) && $(CARGO) build

run-java-server1:
	$(JAVA) -cp $(JAVA_BIN_DIR) $(SERVER) 45914

run-java-server2:
	$(JAVA) -cp $(JAVA_BIN_DIR) $(SERVER) 55914

# Clean up build artifacts
clean:
	rm -f $(JAVA_BIN_DIR)/*.class
	cd $(RUST_CLIENT_DIR) && $(CARGO) clean
	cd $(RUST_SERVER_DIR) && $(CARGO) clean
