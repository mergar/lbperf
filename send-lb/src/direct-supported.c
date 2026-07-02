#define _GNU_SOURCE

#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>

//#ifndef O_DIRECT
//#define O_DIRECT	00040000	/* direct disk access hint */
//#endif

int main(int argc, char *argv[])
{
	if (argc!=2) {
		printf("usage: <progname> <path>\n");
		exit(64);		// EX_USAGE
	}

	int fd = open(argv[1], O_RDONLY|O_DIRECT);

	if (fd < 0) {
		printf("open direct failed: %s\n",argv[1]);
		exit(1);
	}

	close(fd);
	exit(0);
}
