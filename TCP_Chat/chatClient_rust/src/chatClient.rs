/* 20195914 */
/* Jeong eui chan */

use std::env;
use std::io::{self, BufRead, BufReader, Write};
use std::net::TcpStream;
use std::sync::mpsc;
use std::sync::mpsc::{Receiver, Sender};
use std::sync::{Arc, Mutex};
use std::thread;
use ctrlc;

fn main() {
    // Get command line arguments
    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        eprintln!("Usage: cargo run <nickname>");
        return;
    }
     // Connect to the server
    let nickname = &args[1];
    let server_address = "nsl2.cau.ac.kr:25914";

      // Clone the stream for use in the reading thread
    let mut stream = TcpStream::connect(server_address).expect("Failed to connect to server");
    println!("Connected to the server at {}", server_address);

    // Send the nickname to the server
    let stream_clone = stream.try_clone().expect("Failed to clone the stream.");

    // Create a channel for message passing between threads
    write_to_stream(&mut stream, &(nickname.clone() + "\n"));

    // Flag to indicate if termination is requested
    let (tx, rx): (Sender<String>, Receiver<String>) = mpsc::channel();
    let rx_clone = Arc::new(Mutex::new(rx));

    let terminate_signal = Arc::new(Mutex::new(false));
    let terminate_signal_clone = Arc::clone(&terminate_signal);

    // Spawn a thread to handle incoming messages from the server
    thread::spawn(move || {
        handle_incoming_messages(stream_clone, tx, terminate_signal_clone);
    });

    // Set up a Ctrl-C handler to handle termination
    ctrlc::set_handler(move || {
        println!("\ngg~");
        std::process::exit(0);
    }).expect("Error setting Ctrl-C handler");

    // Spawn a thread to display received messages
    let rx_clone_for_display = Arc::clone(&rx_clone);
    thread::spawn(move || {
        loop {
            if let Ok(message) = rx_clone_for_display.lock().unwrap().recv() {
                print!("{}", message);
            }
        }
    });

    // Loop to handle user input
    user_input_loop(&mut stream, terminate_signal);
}

// Function to handle incoming messages from the server
fn handle_incoming_messages(mut stream: TcpStream, tx: Sender<String>, terminate_signal: Arc<Mutex<bool>>) {
    let mut reader = BufReader::new(&mut stream);
    loop {
        let mut buffer = String::new();
        match reader.read_line(&mut buffer) {
            Ok(0) => {
                println!("Disconnected from server ");
                *terminate_signal.lock().unwrap() = true;
                std::process::exit(0);
            },
            Ok(_) => {
                if tx.send(buffer).is_err() {
                    println!("Failed to send message to the receiver.");
                    *terminate_signal.lock().unwrap() = true;
                    break;
                }
            },
            Err(e) => {
                println!("Failed to read from stream: {}", e);
                *terminate_signal.lock().unwrap() = true;
                break;
            }
        }
    }
}


// Function to write a message to the server
fn write_to_stream(stream: &mut TcpStream, message: &str) {
    if let Err(e) = stream.write_all(message.as_bytes()) {
        println!("Failed to write to stream: {}", e);
    }
}


// Function to handle user input and send it to the server
fn user_input_loop(stream: &mut TcpStream, terminate_signal: Arc<Mutex<bool>>) {
    let stdin = io::stdin();
    for line in stdin.lock().lines() {
        if *terminate_signal.lock().unwrap() {
            break;
        }
        match line {
            Ok(input) => {
                let trimmed = input.trim();
                if trimmed == "\\quit" {
                    println!("Disconnecting...");
                    break;
                }
                write_to_stream(stream, &(trimmed.to_string() + "\n"));
            },
            Err(_) => println!("Error reading from stdin."),
        }
    }

    // Send a quit message to the server and exit
    write_to_stream(stream, "\\quit\n");
    println!("gg~");
}
