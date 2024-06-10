#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>

int main() {
	unsigned int result;
	int fd;     /* File descriptor */
    int i, j;   /* Loop variables */
	
	char input = 0;
	
    /* open device file for reading and writing */ 
    /* Use 'open' to open '/dev/multiplier' */
	
	fd = open("/dev/multiplier", O_RDWR);
	
	/* Handle error opening file */
    if(fd == -1) {
        printf("Failed to open device file!\n");
        return -1;
    }
    unsigned int read_i;
    unsigned int read_j;
    int RWBuffer[3];
	while (input != 'q') { /* continue unless user entered 'q' */
		for (i = 0; i <= 16; i++) {
			for (j = 0; j <= 16; j++) {
                /* write values to registers using char dev */
                /* use write to write i and j to peripheral */
                RWBuffer[0] = i;
                RWBuffer[1] = j;
                write(fd,(char*)&RWBuffer,2*sizeof(int));//two ints (i and j) so 2 * sizeof int
                /*read i, j and result using chardev*/
                /*use read to read from peripheral*/
                read(fd,(char*)RWBuffer, 3*sizeof(int)); //3 ints (i, j, result) so 3 * sizeof int
                //need to read from buffer now
                read_i = RWBuffer[0];
                read_j = RWBuffer[1];
                result = RWBuffer[2];
				/*print unsigned ints to screen*/
				printf("%u * %u = %u\n", read_i, read_j, result);


				/* validate result */
				if (result == (i * j)) {
					printf("Result Correct!\n");
				}
				else {
					printf("Result Incorrect!\n");
				}
				/* read from terminal */ 
				input = getchar();
				if (input == 'q') break;
			}
			if (input == 'q') break;
		}
	}
	close(fd);
	return 0;
}