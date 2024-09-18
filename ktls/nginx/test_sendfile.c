
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/sendfile.h>
#include <unistd.h>

/*
 * Confirmin sendFile() support
 *  touch source.txt && touch dest.txt
 *  gcc -o test_sendfile test_sendfile.c
 *  ./test_sendfile
 */

int main() {
  int source = open("source.txt", O_RDONLY);
  int dest = open("dest.txt", O_WRONLY | O_CREAT, 0644);

  if (source == -1 || dest == -1) {
    perror("Failed to open files");
    exit(EXIT_FAILURE);
  }

  off_t offset = 0;
  ssize_t bytes_sent = sendfile(dest, source, &offset, 4096);

  if (bytes_sent == -1) {
    perror("sendfile failed");
    exit(EXIT_FAILURE);
  }

  printf("Bytes sent: %ld\n", bytes_sent);

  close(source);
  close(dest);
  return 0;
}
