/* 20195914 Jeong eui chan */
import java.io.*;
import java.net.*;
import java.time.Duration;
import java.util.Scanner;

public class EasyTCPClient {

    /* startMenu displays the menu and returns the user choice */
    private static int startMenu() {
        Scanner scanner = new Scanner(System.in);
        System.out.println("<Menu>");
        System.out.println("1) Convert text to UPPER-case");
        System.out.println("2) Get server running time");
        System.out.println("3) Get my IP address and port number");
        System.out.println("4) Get server request count");
        System.out.println("5) Exit");
        System.out.print("Input option: ");
        return scanner.nextInt();
    }

    public static void main(String[] args) {
        String serverName = "nsl2.cau.ac.kr";
        int serverPort = 25914;

        try (Socket socket = new Socket(serverName, serverPort);
             PrintWriter out = new PrintWriter(socket.getOutputStream(), true);
             BufferedReader in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
             Scanner scanner = new Scanner(System.in)) {

            while (true) {
                int choice = startMenu();

                if (choice == 5) {
                    System.out.println("Bye bye~");
                    break;
                }

                String message = "";
                if (choice == 1) {
                    System.out.print("Input sentence: ");
                    message = "1:" + scanner.nextLine();
                } else {
                    message = String.valueOf(choice);
                }

                long startTime = System.currentTimeMillis();
                out.println(message);

                String response = in.readLine();
                long endTime = System.currentTimeMillis();
                double elapsedTime = endTime - startTime;

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
                    case 4:
                        System.out.println("Reply from server: requests served = " + response);
                        break;
                    default:
                        System.out.println("Reply from server: " + response);
                        break;
                }

                System.out.printf("RTT = %.3f ms\n", elapsedTime);
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
