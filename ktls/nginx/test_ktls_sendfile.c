#include <fcntl.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/sendfile.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

/*
 *  gcc -o test_ktls_sendfile test_ktls_sendfile.c
 *
 *  ./test_ktls_sendfile
 *
 *  nc localhost 8080
 *
 *
 */
int main() {
  int s = socket(AF_INET, SOCK_STREAM, 0);
  int file = open("testfile.txt", O_RDONLY);
  struct sockaddr_in addr;

  if (s < 0 || file < 0) {
    perror("Failed to create socket or open file");
    exit(EXIT_FAILURE);
  }

  addr.sin_family = AF_INET;
  addr.sin_port = htons(8080);
  addr.sin_addr.s_addr = INADDR_ANY;

  if (bind(s, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
    perror("Bind failed");
    exit(EXIT_FAILURE);
  }

  listen(s, 1);
  int client = accept(s, NULL, NULL);

  if (client < 0) {
    perror("Accept failed");
    exit(EXIT_FAILURE);
  }

  off_t offset = 0;
  ssize_t bytes_sent = sendfile(client, file, &offset, 4096);

  if (bytes_sent < 0) {
    perror("sendfile failed");
  } else {
    printf("Bytes sent: %ld\n", bytes_sent);
  }

  close(file);
  close(client);
  close(s);
  return 0;
}
