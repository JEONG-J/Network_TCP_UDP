/* 20195914 Jeong eui chan */

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.*;
import java.time.Duration;
import java.time.Instant;

public class EasyUDPClient {
    public static void main(String[] args) {
        String serverName = "nsl2.cau.ac.kr";  // Define the server address
        int serverPort = 25914;  // Define the server port

        Runtime.getRuntime().addShutdownHook(new Thread(() -> System.out.println("\nBye bye~")));

        /* Creating a UDP socket */
        try (DatagramSocket socket = new DatagramSocket()) {
            InetAddress serverAddr = InetAddress.getByName(serverName);

            /* Print the local port the client is using */
            System.out.println("Client is running on port " + socket.getLocalPort());
            BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));

            while (true) {
                // Display menu and get user's choice
                int choice = startMenu(reader);

                // If choice is 5, break the loop
                if (choice == 5) {
                    System.out.println("Bye bye~");
                    break;
                }

                // Formulate the message to be sent based on the user choice
                String message;
                if (choice == 1) {
                    System.out.print("Input sentence: ");
                    String input = reader.readLine().trim();
                    message = "1:" + input;
                } else {
                    message = Integer.toString(choice);
                }

                // Record rhe stat time, send the packet and wait for a response
                Instant startTime = Instant.now();
                byte[] sendData = message.getBytes();
                DatagramPacket sendPacket = new DatagramPacket(sendData, sendData.length, serverAddr, serverPort);
                socket.send(sendPacket);

                byte[] receiveData = new byte[1024];
                DatagramPacket receivePacket = new DatagramPacket(receiveData, receiveData.length);
                socket.receive(receivePacket);
                String response = new String(receivePacket.getData(), 0, receivePacket.getLength()).trim();

                // Calculate and display the round-trip time
                Instant endTime = Instant.now();
                long durationInMillis = endTime.toEpochMilli() - startTime.toEpochMilli();

                handleResponse(choice, response);
                System.out.printf("RTT = %.3f ms\n", (float) durationInMillis);
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    /* Method to display a menu to the user and get choice */
    private static int startMenu(BufferedReader reader) throws IOException {
        System.out.println("<Menu>");
        System.out.println("1) Convert text to UPPER-case");
        System.out.println("2) Get server running time");
        System.out.println("3) Get my IP address and port number");
        System.out.println("4) Get server request count");
        System.out.println("5) Exit");
        System.out.print("Input option: ");
        return Integer.parseInt(reader.readLine());
    }

    /* Method to handle the server response based on the user choice */
    private static void handleResponse(int choice, String response) {
        switch (choice) {
            case 2:
                String[] timeParts = response.split(":");
                if (timeParts.length == 3) {
                    String formattedTime = timeParts[0] + ":" + timeParts[1] + ":" + timeParts[2];
                    System.out.println("Reply from server: run time = " + formattedTime);
                } else {
                    System.out.println("Invalid time format received");
                }
                break;
            case 3:
                String[] parts = response.split(":");
                if (parts.length == 2) {
                    System.out.println("Reply from server: client IP = " + parts[0] + ", port = " + parts[1]);
                } else {
                    break;
                }
                break;
            case 4:
                System.out.println("Reply from server: requests served = " + response);
                break;
            default:
                System.out.println("Reply from server: " + response);
                break;
        }
    }
}
