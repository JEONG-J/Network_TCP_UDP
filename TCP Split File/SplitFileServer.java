/* 20195914 */
/* Jeong eui chan */

import java.io.*;
import java.net.*;

public class SplitFileServer {
    public static void main(String[] args) {
        // Check if the correct number of arguments is provided
        if (args.length != 1) {
            System.out.println("Usage: java SplitFileServer <port>");
            return;
        }

        int port = Integer.parseInt(args[0]);
        // Create a server socket to listen on the specified port
        try (ServerSocket serverSocket = new ServerSocket(port)) {
            System.out.println("Server is listening on port " + port);

            // Continuously accept new client connections
            while (true) {
                Socket socket = serverSocket.accept();
                // Create a new thread to handle each client connection
                new ServerThread(socket).start();
            }
        } catch (IOException ex) {
            System.out.println("Server exception: " + ex.getMessage());
            ex.printStackTrace();
        }
    }
}

class ServerThread extends Thread {
    private Socket socket;

    public ServerThread(Socket socket) {
        this.socket = socket;
    }

    public void run() {
        try (InputStream input = socket.getInputStream();
             BufferedReader reader = new BufferedReader(new InputStreamReader(input));
             OutputStream output = socket.getOutputStream();
             PrintWriter writer = new PrintWriter(output, true)) {

            // Read the command from the client
            String command = reader.readLine();
            String[] parts = command.split(" ");
            String operation = parts[0];
            String filename = parts[1];

            // Execute the appropriate operation based on the command
            switch (operation) {
                case "put":
                    receiveFile(reader, filename);
                    break;
                case "get":
                    sendFile(writer, output, filename);
                    break;
                default:
                    System.out.println("Unknown command: " + operation);
            }
        } catch (IOException ex) {
            System.out.println("Server exception: " + ex.getMessage());
            ex.printStackTrace();
        }
    }

    // Method to receive a file from the client and save it
    private void receiveFile(BufferedReader reader, String filename) throws IOException {
        try (BufferedWriter fileWriter = new BufferedWriter(new FileWriter(filename))) {
            String line;
            while ((line = reader.readLine()) != null) {
                fileWriter.write(line);
                fileWriter.newLine();
            }
        }
        System.out.println("File received and saved as " + filename);
    }

    // Method to read a file and send it to the client
    private void sendFile(PrintWriter writer, OutputStream output, String filename) throws IOException {
        File file = new File(filename);
        if (!file.exists()) {
            writer.println("Error: file " + filename + " does not exist");
            return;
        }

        try (BufferedReader fileReader = new BufferedReader(new FileReader(file))) {
            String line;
            while ((line = fileReader.readLine()) != null) {
                writer.println(line);
            }
        }
        System.out.println("File sent: " + filename);
    }
}
