/* 20195914 Jeong eui chan */

import java.io.*;
import java.net.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.time.Duration;
import java.time.Instant;

public class EasyTCPServer {
    private static Instant startTime;
    private static AtomicInteger requestCount = new AtomicInteger(0);

    public static void main(String[] args) throws IOException {
        startTime = Instant.now();
        int serverPort = 25914;
        ServerSocket serverSocket = new ServerSocket(serverPort);
        System.out.println("Server is ready to receive on port " + serverPort);

        /* Handle graceful shutdown */
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            System.out.println("\nBye bye~");
        }));

        try {
            while (true) {
                Socket clientSocket = serverSocket.accept();
                new Thread(() -> handleConnection(clientSocket)).start();
            }
        } finally {
            serverSocket.close();
        }
    }

    private static void handleConnection(Socket clientSocket) {
        try {
            BufferedReader in = new BufferedReader(new InputStreamReader(clientSocket.getInputStream()));
            PrintWriter out = new PrintWriter(clientSocket.getOutputStream(), true);

            while (true) {
                String clientMessage = in.readLine();
                if (clientMessage == null) break;

                requestCount.incrementAndGet();
                System.out.println("Connection request from " + clientSocket.getRemoteSocketAddress());


                String[] parts = clientMessage.split(":");
                int command = Integer.parseInt(parts[0]);

                System.out.println("Command " + command);
                String response = "";

                switch (command) {
                    case 1: // Convert text to uppercase
                        if (parts.length > 1) {
                            response = parts[1].toUpperCase();
                        }
                        break;
                    case 2: // Get server running time
                        Duration elapsedTime = Duration.between(startTime, Instant.now());
                        long hours = elapsedTime.toHours();
                        long minutes = elapsedTime.toMinutes() % 60;
                        long seconds = elapsedTime.getSeconds() % 60;
                        response = String.format("%02d:%02d:%02d", hours, minutes, seconds);
                        break;
                    case 3: // Get client IP address and port number
                        InetAddress clientAddress = clientSocket.getInetAddress();
                        int clientPort = clientSocket.getPort();
                        response = clientAddress.getHostAddress() + ":" + clientPort;
                        break;
                    case 4: // Get server request count
                        response = String.valueOf(requestCount.get());
                        break;
                }

                out.println(response);
            }
        } catch (IOException e) {
        } finally {
            try {
                clientSocket.close();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }
}

