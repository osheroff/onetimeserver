/* ex: set noexpandtab: */
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <libgen.h>
#include <ctype.h>

#define ONETIMESERVER_BINARY "onetimeserver-go"
#define TMPFILE_TEMPLATE "/tmp/onetimeserver.XXXXXX"
#define JSON_PREFIX "_onetimeserver_json: "

void tee_child(FILE *child_stdout, int debug, off_t *offset) {
	char *line = NULL;
	size_t linecap = 0;
	ssize_t linelen;

	/* clear EOF */
	fseek(child_stdout, *offset, SEEK_SET);
	while ((linelen = getline(&line, &linecap, child_stdout)) > 0) {
		if ( strncmp(line, JSON_PREFIX, strlen(JSON_PREFIX)) == 0 ) {
			fwrite(line + strlen(JSON_PREFIX), linelen - strlen(JSON_PREFIX), 1, stdout);
			exit(0);
		} else if ( debug ) {
			fwrite(line, linelen, 1, stderr);
		}
		*offset += linelen;
	}

	sleep(1);
}


void exec_child(int new_stdout, char *tmpfile, int argc, char **argv)
{
	int i, j;
	char **new_argv, *dname = dirname(strdup(argv[0]));

	new_argv = malloc(sizeof(char *) * (argc + 3));

	new_argv[0] = malloc(strlen(dname) + strlen(ONETIMESERVER_BINARY) + 1);
	sprintf(new_argv[0], "%s/%s", dname, ONETIMESERVER_BINARY);

	new_argv[1] = "-output";
	new_argv[2] = tmpfile;

	for(i = 1; i < argc; i++)
		new_argv[i + 2] = argv[i];

	new_argv[i + 2] = NULL;

	dup2(new_stdout, STDOUT_FILENO);
	dup2(new_stdout, STDERR_FILENO);

	execv(new_argv[0], new_argv);

	perror("Couldn't execute " ONETIMESERVER_BINARY);
}


/* a teensy bit of C glue overcome go's reluctance to fork() */
int main(int argc, char **argv)
{
	int child, i, child_alive = 1;
	int child_stdout_fd = 0;
	int debug = 0;
	FILE *child_file = NULL;
	char tmpbuf[sizeof(TMPFILE_TEMPLATE) + 1];

	strcpy(tmpbuf, TMPFILE_TEMPLATE);
	child_stdout_fd = mkstemp(tmpbuf);

	if ( !child_stdout_fd ) {
		perror("Couldn't open tempfile: ");
		abort();
	}

	for (i=0 ; i < argc; i++) {
		if ( strcmp(argv[i], "-debug") == 0 )
			debug = 1;
	}

	if ( (child = fork()) ) {
		child_file = fdopen(child_stdout_fd, "r");
		off_t child_offset = 0;
		while ( 1 ) {
			tee_child(child_file, debug, &child_offset);
			if ( wait4(child, NULL, WNOHANG, NULL) != 0 ) {
				tee_child(child_file, debug, &child_offset);

				/* if tee_child didn't exit, we never got booted: true */
				fprintf(stderr, "Child exited without printing info!\n");
				exit(1);
			}
		}
	} else {
		exec_child(child_stdout_fd, tmpbuf, argc, argv);
	}
}
