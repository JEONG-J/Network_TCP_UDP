/* 20195914 */
/* Jeong eui chan */

use std::collections::HashMap;
use std::io::{BufRead, BufReader, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Instant;

// Alias for the type of the client list, which is shared between threads
type ClientList = Arc<Mutex<HashMap<String, (TcpStream, String)>>>;

// Function to handle each client's connection
fn handle_client(mut stream: TcpStream, clients: ClientList) {
    let peer_addr = stream.peer_addr().unwrap().to_string();
    let mut reader = BufReader::new(stream.try_clone().expect("Failed to clone stream"));
    let mut nickname = String::new();

    // Attempt to read the nickname sent by the client
    if let Ok(_) = reader.read_line(&mut nickname) {
        nickname = nickname.trim().to_string();
        let mut clients_guard = clients.lock().unwrap();
        clients_guard.insert(nickname.clone(), (stream.try_clone().expect("Failed to clone stream"), peer_addr.clone()));
        let num_clients = clients_guard.len();
        drop(clients_guard);

        // Broadcast the join message to all clients
        let server_ip_port = "nsl2.cau.ac.kr:25914";
        let welcome_message = format!(
            "welcome {} to CAU net-class chat room at {}. There are {} users in the room\n",
            nickname, server_ip_port, num_clients
        );
        stream.write_all(welcome_message.as_bytes()).expect("Failed to send welcome message");
        stream.flush().expect("Failed to flush stream");

        let join_message = format!(
            "{} has joined the chat. There are {} users now.\n",
            nickname, num_clients
        );
        broadcast_message(&clients, &join_message, &nickname);

        // Log the client's join action on the server
        let server_side_message = format!(
            "{} joined from {}. There are {} users in the room",
            nickname, peer_addr, num_clients
        );
        println!("{}", server_side_message);

        // Main loop for handling client messages
        loop {
            let mut message = String::new();
            // Read a line from the client
            if let Ok(bytes_read) = reader.read_line(&mut message) {
                if bytes_read == 0 {
                    break;
                }
                message = message.trim().to_string();

                if message.to_lowercase().contains("i hate professor") {
                    let kick_message = format!("{}, Disconnected chatRoom.\n", nickname);
                    broadcast_message(&clients, &kick_message, &nickname);
                    println!("{}", kick_message);

                    let goodbye_message = format!("{} has left the room. There are {} users now.", nickname, clients.lock().unwrap().len() - 1);
                    stream.write_all(goodbye_message.as_bytes()).expect("Failed to send goodbye message");
                    stream.flush().expect("Failed to flush stream");

                    break;
                }

                if message.starts_with("\\") {
                    handle_command(&nickname, &message, &clients);
                } else {
                    let formatted_message = format!("{}: {}\n", nickname, message);
                    broadcast_message(&clients, &formatted_message, &nickname);
                }
            } else {
                break;
            }
        }

        // Remove the client from the client list upon disconnection
        let mut clients_guard = clients.lock().unwrap();
        if clients_guard.remove(&nickname).is_some() {
            let goodbye_message = format!("{} has left the room. There are {} users now.", nickname, clients_guard.len());
            drop(clients_guard);
            broadcast_message(&clients, &goodbye_message, &nickname);
            println!("{}", goodbye_message); // Print on server
        }
    }
}

// Function to handle client commands
fn handle_command(nickname: &str, message: &str, clients: &ClientList) {
    let parts: Vec<&str> = message.splitn(3, ' ').collect();
    let command = parts[0];

    match command {
        "\\ls" => list_clients(nickname, clients),
        "\\secret" => {
            if parts.len() < 3 || !client_exists(parts[1], clients) {
                send_message(nickname, "Error: Nickname does not exist.\n", clients);
                return;
            }
            send_secret(nickname, parts[1], parts[2], clients);
        },
        "\\except" => {
            if parts.len() < 3 {
                send_message(nickname, "Usage: \\except <nickname> <message>\n", clients);
                return;
            }
            if !client_exists(parts[1], clients) {
                send_message(nickname, "Error: Nickname does not exist.\n", clients);
                return;
            }
            broadcast_except(nickname, parts[1], parts[2], clients);
        },
        "\\ping" => send_ping(nickname, clients),
        "\\quit" => {
            let mut clients_guard = clients.lock().unwrap();
            clients_guard.remove(nickname);
            let goodbye_message = format!("{} has left the chat. There are {} users now.", nickname, clients_guard.len());
            drop(clients_guard);
            broadcast_message(clients, &goodbye_message, nickname);
            println!("{}", goodbye_message); // Print on server
        },
        _ => {
            send_message(nickname, "Invalid command\n", clients);
            println!("Invalid command: {}", message);
        }
    }
}

// Function to check if a client exists in the client list
fn client_exists(nickname: &str, clients: &ClientList) -> bool {
    let clients_guard = clients.lock().unwrap();
    clients_guard.contains_key(nickname)
}

// Function to send a message to a specific client
fn send_message(nickname: &str, message: &str, clients: &ClientList) {
    let mut clients_guard = clients.lock().unwrap();
    if let Some((ref mut stream, _)) = clients_guard.get_mut(nickname) {
        stream.write_all(message.as_bytes()).expect("Failed to send message");
        stream.flush().expect("Failed to flush stream");
    }
}

// Function to list all connected clients and send the list to the requesting client
fn list_clients(nickname: &str, clients: &ClientList) {

    let clients_guard = clients.lock().unwrap();
    let client_info: Vec<(String, String)> = clients_guard.iter()
        .map(|(nick, (_, addr))| (nick.clone(), addr.clone()))
        .collect();
    drop(clients_guard); 

    let mut response = String::new();
    response.push_str("Connected users:\n");
    for (nick, addr) in client_info {
        response.push_str(&format!("{} - {}\n", nick, addr));
    }

    send_message(nickname, &response, clients);
}

// Function to send a private message from one client to another
fn send_secret(sender: &str, receiver: &str, message: &str, clients: &ClientList) {
    let mut clients_guard = clients.lock().unwrap();
    if let Some((ref mut stream, _)) = clients_guard.get_mut(receiver) {
        let formatted_message = format!("from {}: {}\n", sender, message);
        stream.write_all(formatted_message.as_bytes()).expect("Failed to send secret message");
        stream.flush().expect("Failed to flush stream");
    }
}

// Function to broadcast a message to all clients except the sender and a specified excluded client
fn broadcast_except(sender: &str, exclude: &str, message: &str, clients: &ClientList) {
    let formatted_message = format!("{}: {}\n", sender, message);
    let mut clients_guard = clients.lock().unwrap();
    for (nick, (ref mut stream, _)) in clients_guard.iter_mut() {
        if nick != sender && nick != exclude {
            stream.write_all(formatted_message.as_bytes()).expect("Failed to send except message");
            stream.flush().expect("Failed to flush stream");
        }
    }
}

// Function to calculate round-trip time (RTT) and send it as a message to the requesting client
fn send_ping(nickname: &str, clients: &ClientList) {
    let start = Instant::now();
    let duration = start.elapsed();
    let rtt_message = format!("RTT: {:.3} ms\n", duration.as_secs_f64() * 1000.0);
    send_message(nickname, &rtt_message, clients);
}

fn broadcast_message(clients: &ClientList, message: &str, sender: &str) {
    let clients_guard = clients.lock().unwrap();
    let message = message.to_string(); // Copy this for use within the closure.

    for (nick, (stream, _)) in clients_guard.iter() {
        if nick != sender {
            let mut stream = stream.try_clone().expect("Failed to clone stream");

            // Processed in a separate thread to send messages asynchronously to each client.
            let message = message.clone();
            thread::spawn(move || {
                stream.write_all(message.as_bytes()).expect("Failed to send message");
                stream.flush().expect("Failed to flush stream");
            });
        }
    }
}

fn main() {
    let listener = TcpListener::bind("0.0.0.0:25914").expect("Could not bind");
    println!("Server is running on port 25914");

    let clients: ClientList = Arc::new(Mutex::new(HashMap::new()));

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                let clients = clients.clone();
                thread::spawn(move || {
                    handle_client(stream, clients);
                });
            }
            Err(e) => {
                eprintln!("Failed to accept connection: {}", e);
            }
        }
    }
}
