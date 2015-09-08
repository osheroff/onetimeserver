#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>

#define ONETIMESERVER_BINARY "onetimeserver.go"
#define TMPFILE_TEMPLATE "/tmp/onetimeserver.XXXXXX"
#define BOOTED_STRING "booted: true"

int tee_child(FILE *child_stdout) {
	char *line = NULL;
	size_t linecap = 0;
	ssize_t linelen;

	/* clear EOF */
	fseek(child_stdout, 0, SEEK_CUR);
	while ((linelen = getline(&line, &linecap, child_stdout)) > 0) {
		if ( strncmp(line, BOOTED_STRING, strlen(BOOTED_STRING)) == 0 )
			exit(0);

		fwrite(line, linelen, 1, stdout);
	}

	sleep(1);

	return 0;
}


void exec_child(int new_stdout, int argc, char **argv)
{
	int i, j;
	char **new_argv, ppidbuf[8];

	sprintf(ppidbuf, "%d", getppid());
	new_argv = malloc(sizeof(char *) * (argc + 4));

	new_argv[0] = argv[1];
	new_argv[1] = "-ppid";
	new_argv[2] = ppidbuf;
	new_argv[3] = "--";

	/* skip argv[0] (wrapper), and argv[1] (program name) */
	for(i = 2, j = 4; i < argc ; i++, j++)
		new_argv[j] = argv[i];

	new_argv[j] = NULL;

	dup2(new_stdout, STDOUT_FILENO);
	dup2(new_stdout, STDERR_FILENO);
	execv(new_argv[0], new_argv);

}


/* a teensy bit of C glue overcome go's reluctance to fork() */
int main(int argc, char **argv)
{
	int child, child_alive = 1;
	int child_stdout_fd = 0;
	FILE *child_file = NULL;
	char tmpbuf[sizeof(TMPFILE_TEMPLATE) + 1];

	strcpy(tmpbuf, TMPFILE_TEMPLATE);
	child_stdout_fd = mkstemp(tmpbuf);

	if ( !child_stdout_fd ) {
		perror("Couldn't open tempfile: ");
		abort();
	}

	printf("output: %s\n", tmpbuf);


	if ( (child = fork()) ) {
		child_file = fdopen(child_stdout_fd, "r");
		while ( 1 ) {
			tee_child(child_file);
			if ( wait4(child, NULL, WNOHANG, NULL) != 0 ) {
				tee_child(child_file);

				/* if tee_child didn't exit, we never got booted: true */
				fprintf(stderr, "Child exited without printing info!\n");
				exit(1);
			}
		}
	} else {
		exec_child(child_stdout_fd, argc, argv);
	}
}
