/* 20195914 */
/* Jeong eui chan */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/select.h>
#include <sys/time.h>
#include <ctype.h>

#define PORT 25914
#define BUF_SIZE 1024
#define MAX_CLIENTS 8
#define NICKNAME_LEN 32

typedef struct {
    int sock;
    char nickname[NICKNAME_LEN];
    struct sockaddr_in addr;
} Client;

Client clients[MAX_CLIENTS];
int client_count = 0;
fd_set reads, temps;
int fd_max;

void send_message(char *message, int len, int sender_sock);
void error_handling(char *message);
int is_nickname_used(char *nickname);
void trim_newline(char *str);
Client *get_client_by_sock(int sock);
Client *get_client_by_nickname(char *nickname);
void handle_command(int sender_sock, char *command);
void list_users(int sender_sock);
void send_secret_message(int sender_sock, char *nickname, char *message);
void send_except_message(int sender_sock, char *nickname, char *message);
void handle_ping(int sender_sock);
void handle_quit(int sender_sock);
void remove_client(int sock);
char *strcasestr(const char *haystack, const char *needle);

int main(void) {
    int serv_sock, clnt_sock;
    struct sockaddr_in serv_addr, clnt_addr;
    socklen_t clnt_addr_size;
    int fd_num, str_len;
    char message[BUF_SIZE];
    struct timeval timeout;

      // Create server socket
    serv_sock = socket(PF_INET, SOCK_STREAM, 0);
    if (serv_sock == -1)
        error_handling("socket() error");

    // Configure server address
    memset(&serv_addr, 0, sizeof(serv_addr));
    serv_addr.sin_family = AF_INET;
    serv_addr.sin_addr.s_addr = htonl(INADDR_ANY);
    serv_addr.sin_port = htons(PORT);

    // Bind socket to address
    if (bind(serv_sock, (struct sockaddr *)&serv_addr, sizeof(serv_addr)) == -1)
        error_handling("bind() error");

    // Listen for incoming connections
    if (listen(serv_sock, 5) == -1)
        error_handling("listen() error");

    FD_ZERO(&reads);
    FD_SET(serv_sock, &reads);
    fd_max = serv_sock;

    // Infinite loop to handle incoming connections and messages
    while (1) {
        temps = reads;
        timeout.tv_sec = 5;
        timeout.tv_usec = 5000;

        fd_num = select(fd_max + 1, &temps, 0, 0, &timeout);
        if (fd_num == -1)
            break;
        if (fd_num == 0)
            continue;

        for (int i = 0; i < fd_max + 1; i++) {
            if (FD_ISSET(i, &temps)) {
                if (i == serv_sock) {
                    clnt_addr_size = sizeof(clnt_addr);
                    clnt_sock = accept(serv_sock, (struct sockaddr *)&clnt_addr, &clnt_addr_size);
                    if (clnt_sock == -1)
                        error_handling("accept() error");

                    if (client_count >= MAX_CLIENTS) {
                            write(clnt_sock, "Chatting room full. Cannot connect\n", 36);
                            close(clnt_sock);
                        } else {
                            str_len = read(clnt_sock, message, NICKNAME_LEN);
                            message[str_len] = 0;
                            trim_newline(message);

                            if (is_nickname_used(message)) {
                                write(clnt_sock, "Nickname already used by another user. Cannot connect\n", 55);
                                close(clnt_sock);
                            } else {
                                FD_SET(clnt_sock, &reads);
                                if (clnt_sock > fd_max)
                                    fd_max = clnt_sock;

                                clients[client_count].sock = clnt_sock;
                                strncpy(clients[client_count].nickname, message, NICKNAME_LEN);
                                clients[client_count].addr = clnt_addr;
                                client_count++;

                                // Constructing the welcome message
                                char welcome_msg[BUF_SIZE];
                                snprintf(welcome_msg, sizeof(welcome_msg), "Welcome %s to CAU net-class chat room at %s:%d. There are %d users in the room\n",
                                        clients[client_count - 1].nickname, inet_ntoa(clients[client_count - 1].addr.sin_addr), ntohs(clients[client_count - 1].addr.sin_port), client_count);

                                // Send welcome message to the new client only
                                write(clnt_sock, welcome_msg, strlen(welcome_msg));

                                // Broadcast join message to all other clients
                                char join_msg[BUF_SIZE];
                                snprintf(join_msg, sizeof(join_msg), "%s joined the room. There are %d users now.\n", clients[client_count - 1].nickname, client_count);
                                send_message(join_msg, strlen(join_msg), clnt_sock);

                                // Log the connection on the server console
                                printf("%s joined from %s:%d. There are %d users in the room\n",
                                    clients[client_count - 1].nickname,
                                    inet_ntoa(clients[client_count - 1].addr.sin_addr),
                                    ntohs(clients[client_count - 1].addr.sin_port),
                                    client_count);
                            }
                        }
                } else {
                    str_len = read(i, message, BUF_SIZE);
                    if (str_len == 0) {
                    // If the client terminates the connection
                        remove_client(i);
                    } else {
                        message[str_len] = '\0';
                        trim_newline(message); // Remove the newline character at the end of the message.

                        // When the user enters a command
                        if (str_len > 1 && message[0] == '\\') {
                            handle_command(i, message);
                        } else {
                            // Check if message is not empty
                            if(strlen(message) > 0) {
                                Client *sender = get_client_by_sock(i);
                                char chat_message[BUF_SIZE + NICKNAME_LEN];
                                snprintf(chat_message, sizeof(chat_message), "%s> %s\n", sender->nickname, message);

                                // Disconnected when user writes "i hate professor"
                                if (strcasestr(message, "I hate professor")) {
                                    remove_client(i);
                                } else {
                                    send_message(chat_message, strlen(chat_message), i);
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    close(serv_sock);
    return 0;
}

void send_message(char *message, int len, int sender_sock) {
    for (int i = 0; i < client_count; i++) {
        if (clients[i].sock != sender_sock) {
            write(clients[i].sock, message, len);
        }
    }
}

// check if the given nickname is aleready used by any client
int is_nickname_used(char *nickname) {
    for (int i = 0; i < client_count; i++) {
        if (strcmp(clients[i].nickname, nickname) == 0) {
            return 1;
        }
    }
    return 0;
}

// Remove the newline character from the end of the string, if present
void trim_newline(char *str) {
    int len = strlen(str);
    if (str[len - 1] == '\n') {
        str[len - 1] = '\0';
    }
}

Client *get_client_by_sock(int sock) {
    for (int i = 0; i < client_count; i++) {
        if (clients[i].sock == sock) {
            return &clients[i];
        }
    }
    return NULL;
}

Client *get_client_by_nickname(char *nickname) {
    for (int i = 0; i < client_count; i++) {
        if (strcmp(clients[i].nickname, nickname) == 0) {
            return &clients[i];
        }
    }
    return NULL;
}

void handle_command(int sender_sock, char *command) {
    struct timeval start, end;
    gettimeofday(&start, NULL);
    if (strncmp(command, "\\ls", 3) == 0) {
        // List all connected users
        list_users(sender_sock);
    } else if (strncmp(command, "\\secret ", 8) == 0) {
        // Send a secret message to a specified user
        char *nickname = command + 8;
        char *message = strchr(nickname, ' ');
        if (message != NULL) {
            *message = '\0';
            message++;
            send_secret_message(sender_sock, nickname, message);
        } else {
            write(sender_sock, "Invalid command\n", 17);
        }
    } else if (strncmp(command, "\\except ", 8) == 0) {
        // Send a message to all users except one specified user
        char *nickname = command + 8;
        char *message = strchr(nickname, ' ');
        if (message != NULL) {
            *message = '\0'; 
            message++; 
            trim_newline(message);
            send_except_message(sender_sock, nickname, message);
        } else {
            write(sender_sock, "Invalid command\n", 17);
        }
    } else if (strncmp(command, "\\ping", 5) == 0) {
        // Calculate and send the round-trip time (RTT) for a ping command
        gettimeofday(&end, NULL);
        double rtt = ((end.tv_sec - start.tv_sec) * 1000.0) + ((end.tv_usec - start.tv_usec) / 1000.0);
        char ping_response[BUF_SIZE];
        snprintf(ping_response, BUF_SIZE, "RTT: %.3f ms\n", rtt);
        write(sender_sock, ping_response, strlen(ping_response));
    } else if (strncmp(command, "\\quit", 5) == 0) {
        // handle quit command
        handle_quit(sender_sock); 
    } else {
        char invalid_command_message[BUF_SIZE];
        snprintf(invalid_command_message, sizeof(invalid_command_message), "Invalid command: %s\n", command);
        write(sender_sock, invalid_command_message, strlen(invalid_command_message));
        printf("Invalid command: %s\n", command);
    }
}

// Send a list of all connected users to the requesting client
void list_users(int sender_sock) {
    char message[BUF_SIZE];
    int len = 0;
    len += snprintf(message + len, sizeof(message) - len, "List of connected users:\n");
    for (int i = 0; i < client_count; i++) {
        len += snprintf(message + len, sizeof(message) - len, "<%s, %s, %d>\n", clients[i].nickname,
                        inet_ntoa(clients[i].addr.sin_addr), ntohs(clients[i].addr.sin_port));
    }
    write(sender_sock, message, len);
}

// Send a secret message to a specific user
void send_secret_message(int sender_sock, char *nickname, char *message) {
    Client *target = get_client_by_nickname(nickname);
    if (target != NULL) {
        char secret_message[BUF_SIZE];
        snprintf(secret_message, sizeof(secret_message), "from %s > %s\n", get_client_by_sock(sender_sock)->nickname, message);
        write(target->sock, secret_message, strlen(secret_message));
    } else {
        write(sender_sock, "Nickname not found\n", 20);
    }
}

// Send a message to all users except one specified user
void send_except_message(int sender_sock, char *nickname, char *message) {
    Client *target = get_client_by_nickname(nickname);
    if (target != NULL) {
        for (int i = 0; i < client_count; i++) {
            if (clients[i].sock != target->sock && clients[i].sock != sender_sock) {
                write(clients[i].sock, message, strlen(message)); // 메시지만 전송
            }
        }
    } else {
        write(sender_sock, "Nickname not found\n", 20);
    }
}


void handle_quit(int sender_sock) {
    remove_client(sender_sock);
}

// Remove a client from the list and notify others
void remove_client(int sock) {
    for (int i = 0; i < client_count; i++) {
        if (clients[i].sock == sock) {
            char message[BUF_SIZE];
            snprintf(message, sizeof(message), "%s left the room. There are %d users now.\n", clients[i].nickname, client_count - 1);
            send_message(message, strlen(message), -1);

            printf("%s left the room. There are %d users now.\n", clients[i].nickname, client_count - 1);

            close(sock);
            FD_CLR(sock, &reads);

            for (int j = i; j < client_count - 1; j++)
                clients[j] = clients[j + 1];
            client_count--;
            break;
        }
    }
}

void error_handling(char *message) {
    fputs(message, stderr);
    fputc('\n', stderr);
    exit(1);
}

// Define strcasestr if not available
char *strcasestr(const char *haystack, const char *needle) {
    if (!*needle)
        return (char *)haystack;
    for (; *haystack; haystack++) {
        if (toupper((unsigned char)*haystack) == toupper((unsigned char)*needle)) {
            const char *h, *n;
            for (h = haystack, n = needle; *h && *n; h++, n++) {
                if (toupper((unsigned char)*h) != toupper((unsigned char)*n)) {
                    break;
                }
            }
            if (!*n) {
                return (char *)haystack;
            }
        }
    }
    return NULL;
}
