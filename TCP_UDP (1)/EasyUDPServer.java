/* 20195914 Jeong eui chan */

import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;
import java.time.Duration;
import java.time.Instant;
import java.util.StringTokenizer;


public class EasyUDPServer {
    public static void main(String[] args) {
        int serverPort = 25914; // Define the server port
        Instant startTime = Instant.now(); // Record the start time for uptime calculation
        int requestCount = 0; // Counter for the number of request

        // Add a shutdown hook to handle "Ctrl-C" interruption
        Runtime.getRuntime().addShutdownHook(new Thread(() -> System.out.println("\nBye bye~")));

        try (DatagramSocket socket = new DatagramSocket(serverPort)) {
            System.out.println("Server is ready to receive on port " + serverPort);

            byte[] buffer = new byte[1024]; // Buffer to store incoming data

            while(true) {
                DatagramPacket request = new DatagramPacket(buffer, buffer.length);
                socket.receive(request);
                requestCount++;

                // Extract the client address and port
                InetAddress clientAddress = request.getAddress();
                int clientPort = request.getPort();
                System.out.printf("Connection request from %s:%d\n", clientAddress.getHostAddress(), clientPort);

                // Extract the message and command from the received packet
                String receiveMsg = new String(request.getData(), 0, request.getLength()).trim();
                StringTokenizer st = new StringTokenizer(receiveMsg, ":");
                int command = Integer.parseInt(st.nextToken());
                System.out.printf("Command %d\n", command);

                String responseStr = "";
                switch (command) {
                    case 1:
                        // Convert received message to uppercase if command is 1
                        if (st.hasMoreTokens()) {
                            responseStr = st.nextToken().toUpperCase();
                        }
                        break;
                    case 2: // Get server running time
                        Duration elapsedTime = Duration.between(startTime, Instant.now());
                        long hours = elapsedTime.toHours();
                        long minutes = elapsedTime.toMinutes() % 60;
                        long seconds = elapsedTime.getSeconds() % 60;
                        responseStr = String.format("%02d:%02d:%02d", hours, minutes, seconds);
                    break;

                    case 3:
                        // Send client's address and port if command is 3
                        responseStr = clientAddress.getHostAddress() + ":" + clientPort;
                        break;
                    case 4:
                        // Send request count if command is 4
                        responseStr = String.valueOf(requestCount);
                        break;
                    default:
                        // No action for unhandled commands
                        break;
                }

                // send a response if response string is not empty
                if (!responseStr.isEmpty()) {
                    byte[] response = responseStr.getBytes();
                    DatagramPacket reply = new DatagramPacket(response, response.length, request.getAddress(), request.getPort());
                    socket.send(reply);
                }
            }
        } catch (SocketException e) {
            System.err.println("SocketException: " + e.getMessage());
        } catch (Exception e) {
            System.err.println("Exception:" + e.getMessage());
        }
    }
}