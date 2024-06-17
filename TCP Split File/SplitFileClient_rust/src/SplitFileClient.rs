/* 20195914 */
/* Jeong eui chan */

use std::env;
use std::fs::File;
use std::io::{Read, Write, BufReader, BufWriter};
use std::net::TcpStream;

const SERVER1_ADDRESS: &str = "nsl2.cau.ac.kr:45914";
const SERVER2_ADDRESS: &str = "nsl5.cau.ac.kr:55914";

// Function to send data to the server
fn send_to_server(address: &str, command: &str, data: &[u8]) -> bool {
    // Connect to the server
    let stream = match TcpStream::connect(address) {
        Ok(stream) => stream,
        Err(e) => {
            eprintln!("Error connecting to server: {}", e);
            return false;
        }
    };

    let mut writer = BufWriter::new(&stream);
    // Send the command to the server
    if writeln!(writer, "{}", command).is_err() {
        eprintln!("Error sending command");
        return false;
    }

    // Send the data to the server
    if writer.write_all(data).is_err() {
        eprintln!("Error sending data");
        return false;
    }

    true
}

// Function to receive data from the server
fn receive_from_server(address: &str, command: &str) -> Vec<u8> {
    // Connect to the server
    let stream = TcpStream::connect(address).expect("Error connecting to server");
    let mut writer = BufWriter::new(&stream);
    writeln!(writer, "{}", command).expect("Error sending command");

    let mut reader = BufReader::new(&stream);
    let mut buffer = Vec::new();
    // Read the data from the server
    reader.read_to_end(&mut buffer).expect("Error receiving data");

    // Check for error message from the server
    if buffer.starts_with(b"Error:") {
        eprintln!("{}", String::from_utf8_lossy(&buffer));
        return Vec::new();
    }

    buffer
}

// Function to handle the 'put' command
fn put_file(filename: &str) {
    let mut file = File::open(filename).expect("Error reading file");
    let mut data = Vec::new();
    file.read_to_end(&mut data).expect("Error reading file");

    let (part1, part2) = split_file(&data);

    let part1_filename = format!("{}-part1.txt", filename.trim_end_matches(".txt"));
    let part2_filename = format!("{}-part2.txt", filename.trim_end_matches(".txt"));

    // Send each part to the respective server
    if !send_to_server(SERVER1_ADDRESS, &format!("put {}", part1_filename), &part1) {
        eprintln!("Failed to send part1 to server 1");
        return;
    }

    if !send_to_server(SERVER2_ADDRESS, &format!("put {}", part2_filename), &part2) {
        eprintln!("Failed to send part2 to server 2");
    }
}

// Function to handle the 'get' command
fn get_file(filename: &str) {
    let part1_filename = format!("{}-part1.txt", filename.trim_end_matches(".txt"));
    let part2_filename = format!("{}-part2.txt", filename.trim_end_matches(".txt"));

    // Receive each part from the respective server
    let part1 = receive_from_server(SERVER1_ADDRESS, &format!("get {}", part1_filename));
    if part1.is_empty() {
        eprintln!("Failed to receive part1 from server 1");
        return;
    }

    let part2 = receive_from_server(SERVER2_ADDRESS, &format!("get {}", part2_filename));
    if part2.is_empty() {
        eprintln!("Failed to receive part2 from server 2");
        return;
    }

    // Merge the two parts and save to a new file
    let merged = merge_file(&part1, &part2);
    let merged_filename = format!("{}-merged.txt", filename.trim_end_matches(".txt"));
    let mut file = File::create(merged_filename).expect("Error writing merged file");
    file.write_all(&merged).expect("Error writing merged file");
    println!("Merged file saved as {}-merged.txt", filename.trim_end_matches(".txt"));
}

// Function to split the data into two parts
fn split_file(data: &[u8]) -> (Vec<u8>, Vec<u8>) {
    let mut part1 = Vec::new();
    let mut part2 = Vec::new();
    for (i, &byte) in data.iter().enumerate() {
        if i % 2 == 0 {
            part1.push(byte);
        } else {
            part2.push(byte);
        }
    }
    (part1, part2)
}

// Function to merge two parts into one
fn merge_file(part1: &[u8], part2: &[u8]) -> Vec<u8> {
    let mut merged = Vec::new();
    let mut i = 0;
    let mut j = 0;
    while i < part1.len() || j < part2.len() {
        if i < part1.len() {
            merged.push(part1[i]);
            i += 1;
        }
        if j < part2.len() {
            merged.push(part2[j]);
            j += 1;
        }
    }
    merged
}

// Main function to parse command-line arguments and execute the appropriate command
fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 3 {
        eprintln!("Usage: cargo run -- <put/get> <filename>");
        return;
    }

    let command = &args[1];
    let filename = &args[2];

    match command.as_str() {
        "put" => put_file(filename),
        "get" => get_file(filename),
        _ => eprintln!("Unknown command: {}", command),
    }
}
