/* 20195914 */
/* Jeong eui chan */

use std::env;
use std::fs::File;
use std::io::{Read, Write, BufRead, BufReader};
use std::net::{TcpListener, TcpStream};
use std::thread;

// Function to handle each client connection
fn handle_client(mut stream: TcpStream) {
    let mut reader = BufReader::new(&stream);
    let mut command = String::new();
    
    // Read the command from the client
    if reader.read_line(&mut command).is_err() {
        eprintln!("Failed to read command from client");
        return;
    }

    // Split the command into parts
    let parts: Vec<&str> = command.trim().split_whitespace().collect();
    if parts.len() < 2 {
        eprintln!("Invalid command: {}", command);
        return;
    }

    let command_type = parts[0];
    let filename = parts[1];

    // Execute the appropriate action based on the command type
    match command_type {
        "put" => receive_file(&mut reader, filename),
        "get" => send_file(&mut stream, filename),
        _ => eprintln!("Unknown command: {}", command_type),
    }
}

// Function to receive a file from the client and save it
fn receive_file(reader: &mut BufReader<&TcpStream>, filename: &str) {
    let mut file = File::create(filename).expect("Failed to create file");
    let mut buffer = Vec::new();
    reader.read_to_end(&mut buffer).expect("Failed to read data");
    file.write_all(&buffer).expect("Failed to write data to file");
    println!("File received and saved as {}", filename);
}

// Function to read a file and send it to the client
fn send_file(stream: &mut TcpStream, filename: &str) {
    let mut file = match File::open(filename) {
        Ok(file) => file,
        Err(_) => {
            writeln!(stream, "Error: file {} does not exist", filename).unwrap();
            return;
        }
    };
    let mut buffer = Vec::new();
    file.read_to_end(&mut buffer).expect("Failed to read file");
    stream.write_all(&buffer).expect("Failed to send data");
    println!("File sent: {}", filename);
}

// Main function to start the server and listen for incoming connections
fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        eprintln!("Usage: cargo run <port>");
        return;
    }

    let port = &args[1];
    let listener = TcpListener::bind(format!("0.0.0.0:{}", port)).expect("Failed to bind to port");
    println!("Server is listening on port {}", port);

    // Accept incoming connections and handle them in separate threads
    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                thread::spawn(|| handle_client(stream));
            }
            Err(e) => eprintln!("Failed to accept connection: {}", e),
        }
    }
}
