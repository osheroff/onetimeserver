#!/usr/bin/perl
# a teensy bits of perl glue to overcome go's reluctance to fork()
use File::Temp qw/ tempfile tempdir /;
($fh, $filename) = tempfile();

if ( !($child = fork()) ) {
   open(STDOUT, '>>', $filename) or die "Can't redirect STDOUT: $!";
   open(STDERR, ">&STDOUT")     or die "Can't dup STDOUT: $!";
   exec(@ARGV);
} else {
  print "output: " . $filename . "\n";

  $child_dead = 0;

  open(OUTPUT, '<', $filename);
  for (;;) {
    while (<OUTPUT>) {
      exit(0) if $_ =~ /__done__/;
      print $_;
    }
    if ( $child_dead ) {
      print "error: child died without booting" . "\n";
      exit(1);
    }
    sleep 1;

    seek(OUTPUT, 0, 1);
    if ( kill 0, $child ) {
      # iterate once more after the child is dead so we don't race its exit
      $child_dead = 1;
    }
  }
}





