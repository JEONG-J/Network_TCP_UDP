/**
 * MultiTCPServer.c
 **/

/* 20195914 Jeong eui chan */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <fcntl.h>
#include <errno.h>
#include <ctype.h>
#include <signal.h>

#define INITIAL_CLIENT_CAPACITY 10
#define MAX_BUFFER_SIZE 1024
#define SERVER_PORT 25914

typedef struct {
    int socket_fd;
    int client_id;
    int is_active;
} ClientInfo;

ClientInfo *client_list = NULL;
int client_list_capacity = 0;
int total_clients = 0;
int next_client_id = 1;
int total_requests_handled = 0;
time_t server_start_time, last_print_time;

void handleError(const char *errorMessage) {
    perror(errorMessage);
    exit(1);
}

void setSocketNonBlocking(int socket_fd) {
    int flags = fcntl(socket_fd, F_GETFL);
    if (flags < 0) {
        perror("fcntl(F_GETFL) failed");
        exit(EXIT_FAILURE);
    }
    flags |= O_NONBLOCK;
    if (fcntl(socket_fd, F_SETFL, flags) < 0) {
        perror("fcntl(F_SETFL) failed");
        exit(EXIT_FAILURE);
    }
}

void expandClientList() {
    int new_capacity = client_list_capacity == 0 ? INITIAL_CLIENT_CAPACITY : client_list_capacity * 2;
    client_list = realloc(client_list, new_capacity * sizeof(ClientInfo));
    if (!client_list) {
        perror("Failed to reallocate memory for client list");
        exit(EXIT_FAILURE);
    }
    for (int i = client_list_capacity; i < new_capacity; i++) {
        client_list[i].socket_fd = 0;
        client_list[i].is_active = 0;
    }
    client_list_capacity = new_capacity;
}

void printCurrentTime() {
    char buffer[30];
    time_t now = time(NULL);
    strftime(buffer, sizeof(buffer), "[Time: %H:%M:%S]", localtime(&now));
    printf("%s ", buffer);
}

void printClientStatus(const char *status, int client_id) {
    printCurrentTime();
    printf("Client %d %s. Number of clients connected = %d\n", client_id, status, total_clients);
}

void handleSignalInterrupt(int signal) {
    free(client_list);
    printf("\nBye bye~\n");
    exit(0);
}

void handleClientRequest(int client_index) {
    char buffer[MAX_BUFFER_SIZE];
    int read_bytes = read(client_list[client_index].socket_fd, buffer, MAX_BUFFER_SIZE - 1);
    if (read_bytes == 0) { // Client disconnected
        client_list[client_index].is_active = 0;
        close(client_list[client_index].socket_fd);
        total_clients--;
        printClientStatus("disconnected", client_list[client_index].client_id);
    } else if (read_bytes > 0) {
        buffer[read_bytes] = '\0';
        int command;
        sscanf(buffer, "%d", &command);
        char response[MAX_BUFFER_SIZE];
        switch (command) {
            case 1: { // Convert text to uppercase
                char *text = strchr(buffer, ':') + 1;
                for (int i = 0; text[i]; i++) {
                    text[i] = toupper(text[i]);
                }
                strcpy(response, text);
                break;
            }
            case 2: { // Tell how long the server has been running
                int seconds = difftime(time(NULL), server_start_time);
                int hours = seconds / 3600;
                int minutes = (seconds % 3600) / 60;
                seconds %= 60;
                sprintf(response, "%02d:%02d:%02d", hours, minutes, seconds);
                break;
            }
            case 3: { // Tell the client's IP address and port number
                struct sockaddr_in addr;
                socklen_t addr_len = sizeof(addr);
                getpeername(client_list[client_index].socket_fd, (struct sockaddr *)&addr, &addr_len);
                sprintf(response, "%s:%d", inet_ntoa(addr.sin_addr), ntohs(addr.sin_port));
                break;
            }
            case 4: { // Tell how many requests the server has handled so far
                sprintf(response, "%d", total_requests_handled);
                break;
            }
            default:
                sprintf(response, "Unknown command");
        }
        send(client_list[client_index].socket_fd, response, strlen(response), 0);
        total_requests_handled++;
    }
}

int main() {
    int server_socket, client_socket, port_number;
    struct sockaddr_in server_address, client_address;
    socklen_t client_address_len;
    fd_set active_fd_set;
    struct timeval select_timeout;

    signal(SIGINT, handleSignalInterrupt);

    server_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (server_socket < 0) handleError("ERROR opening socket");
    setSocketNonBlocking(server_socket);

    memset(&server_address, 0, sizeof(server_address));
    port_number = SERVER_PORT;
    server_address.sin_family = AF_INET;
    server_address.sin_addr.s_addr = INADDR_ANY;
    server_address.sin_port = htons(port_number);

    if (bind(server_socket, (struct sockaddr *)&server_address, sizeof(server_address)) < 0) 
        handleError("ERROR on binding");

    printf("Server is ready to receive on port %d\n", port_number);

    listen(server_socket, 5);
    client_address_len = sizeof(client_address);

    expandClientList();  // Initialize client list
    server_start_time = last_print_time = time(NULL);

    while (1) {
        FD_ZERO(&active_fd_set);
        FD_SET(server_socket, &active_fd_set);
        int highest_socket_descriptor = server_socket;

        for (int i = 0; i < total_clients; i++) {
            if (client_list[i].socket_fd > 0 && client_list[i].is_active) {
                FD_SET(client_list[i].socket_fd, &active_fd_set);
                if (client_list[i].socket_fd > highest_socket_descriptor)
                    highest_socket_descriptor = client_list[i].socket_fd;
            }
        }

        select_timeout.tv_sec = 1;
        select_timeout.tv_usec = 0;

        if (select(highest_socket_descriptor + 1, &active_fd_set, NULL, NULL, &select_timeout) < 0) {
            perror("select error");
            continue;
        }

        time_t current_time = time(NULL);
        if (difftime(current_time, last_print_time) >= 10) {
            printCurrentTime();
            printf("Number of clients connected = %d\n", total_clients);
            last_print_time = current_time;
        }

        if (FD_ISSET(server_socket, &active_fd_set)) {
            client_socket = accept(server_socket, (struct sockaddr *)&client_address, &client_address_len);
            if (client_socket < 0) {
                perror("accept failed");
                continue;
            }
            setSocketNonBlocking(client_socket);
            if (total_clients >= client_list_capacity) {
                expandClientList();
            }
            client_list[total_clients].socket_fd = client_socket;
            client_list[total_clients].client_id = next_client_id++;
            client_list[total_clients].is_active = 1;
            total_clients++;
            printClientStatus("connected", client_list[total_clients-1].client_id);
        }

        for (int i = 0; i < total_clients; i++) {
            if (client_list[i].is_active && FD_ISSET(client_list[i].socket_fd, &active_fd_set)) {
                handleClientRequest(i);
            }
        }
    }

    return 0;
}
