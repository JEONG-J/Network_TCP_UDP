/* 20195914 */
/* Jeong eui chan */

import java.io.*;
import java.net.*;

public class SplitFileClient {
    private static final String SERVER1_ADDRESS = "nsl2.cau.ac.kr";
    private static final int SERVER1_PORT = 45914;
    private static final String SERVER2_ADDRESS = "nsl5.cau.ac.kr";
    private static final int SERVER2_PORT = 55914;

    public static void main(String[] args) {
        // Check if the correct number of arguments is provided
        if (args.length != 2) {
            System.out.println("Usage: java SplitFileClient <put/get> <filename>");
            return;
        }

        String command = args[0];
        String filename = args[1];

        try {
            switch (command) {
                case "put":
                    putFile(filename);
                    break;
                case "get":
                    getFile(filename);
                    break;
                default:
                    System.out.println("Unknown command: " + command);
            }
        } catch (IOException ex) {
            System.out.println("Client exception: " + ex.getMessage());
            ex.printStackTrace();
        }
    }

    // Method to handle the 'put' command
    private static void putFile(String filename) throws IOException {
        File file = new File(filename);
        if (!file.exists()) {
            System.out.println("Error: file " + filename + " does not exist");
            return;
        }

        // Read the file and split it into two parts
        byte[] data = readFile(file);
        byte[] part1 = splitFile(data, 0);
        byte[] part2 = splitFile(data, 1);

        String part1Filename = filename.replace(".txt", "-part1.txt");
        String part2Filename = filename.replace(".txt", "-part2.txt");

        // Send each part to the respective server
        if (!sendToServer(SERVER1_ADDRESS, SERVER1_PORT, "put " + part1Filename, part1)) {
            System.out.println("Failed to send part1 to server 1");
            return;
        }
        if (!sendToServer(SERVER2_ADDRESS, SERVER2_PORT, "put " + part2Filename, part2)) {
            System.out.println("Failed to send part2 to server 2");
        }
    }

    // Method to handle the 'get' command
    private static void getFile(String filename) throws IOException {
        String part1Filename = filename.replace(".txt", "-part1.txt");
        String part2Filename = filename.replace(".txt", "-part2.txt");

        // Receive each part from the respective server
        byte[] part1 = receiveFromServer(SERVER1_ADDRESS, SERVER1_PORT, "get " + part1Filename);
        if (part1 == null) {
            System.out.println("Failed to receive part1 from server 1");
            return;
        }

        byte[] part2 = receiveFromServer(SERVER2_ADDRESS, SERVER2_PORT, "get " + part2Filename);
        if (part2 == null) {
            System.out.println("Failed to receive part2 from server 2");
            return;
        }

        // Merge the two parts and save to a new file
        byte[] merged = mergeFile(part1, part2);
        String mergedFilename = filename.replace(".txt", "-merged.txt");
        writeFile(new File(mergedFilename), merged);
        System.out.println("Merged file saved as " + mergedFilename);
    }

    // Method to read a file and return its contents as a byte array
    private static byte[] readFile(File file) throws IOException {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        try (BufferedInputStream bis = new BufferedInputStream(new FileInputStream(file))) {
            byte[] buffer = new byte[1024];
            int bytesRead;
            while ((bytesRead = bis.read(buffer)) != -1) {
                bos.write(buffer, 0, bytesRead);
            }
        }
        return bos.toByteArray();
    }

    // Method to write a byte array to a file
    private static void writeFile(File file, byte[] data) throws IOException {
        try (BufferedOutputStream bos = new BufferedOutputStream(new FileOutputStream(file))) {
            bos.write(data);
        }
    }

    // Method to split a byte array into two parts based on starting index
    private static byte[] splitFile(byte[] data, int start) {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        for (int i = start; i < data.length; i += 2) {
            bos.write(data[i]);
        }
        return bos.toByteArray();
    }

    // Method to merge two byte arrays into one
    private static byte[] mergeFile(byte[] part1, byte[] part2) {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        int i = 0, j = 0;
        while (i < part1.length || j < part2.length) {
            if (i < part1.length) {
                bos.write(part1[i++]);
            }
            if (j < part2.length) {
                bos.write(part2[j++]);
            }
        }
        return bos.toByteArray();
    }

    // Method to send a command and data to a server
    private static boolean sendToServer(String address, int port, String command, byte[] data) {
        try (Socket socket = new Socket(address, port);
             OutputStream output = socket.getOutputStream();
             PrintWriter writer = new PrintWriter(output, true)) {

            writer.println(command);
            output.write(data);
            return true;
        } catch (IOException ex) {
            System.out.println("Error connecting to server: " + ex.getMessage());
            return false;
        }
    }

    // Method to receive data from a server
    private static byte[] receiveFromServer(String address, int port, String command) {
        try (Socket socket = new Socket(address, port);
             InputStream input = socket.getInputStream();
             PrintWriter writer = new PrintWriter(socket.getOutputStream(), true)) {

            writer.println(command);

            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            byte[] buffer = new byte[1024];
            int bytesRead;
            while ((bytesRead = input.read(buffer)) != -1) {
                bos.write(buffer, 0, bytesRead);
            }
            return bos.toByteArray();
        } catch (IOException ex) {
            System.out.println("Error connecting to server: " + ex.getMessage());
            return null;
        }
    }
}
