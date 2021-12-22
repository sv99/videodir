#!/usr/bin/perl
#
#
use strict;
use Pod::Usage;

if (@ARGV < 1)
{
   pod2usage(1);
}

my $dir = shift;

my @files = `find $dir -type f`; # glob( $dir . '/**/*' );

for my $f (@files)
{
    $f =~ s/\s+$//;
    print "$f\n";
    open(FH, '>', $f) or die $!;
    # random string generator - lenght 8 symbols
    my $rr = join '', map { (q(a)..q(z))[rand(26)] } 1..8;
    print FH $rr;
    close(FH);
}

__END__

=for stopwords clean_video_data.pl

=head1 NAME

clean_video_data.pl - video files init with small random data

=head1 SYNOPSIS

 clean_video_data.pl video_data_dir

 Options:
   -h --help           verbose help message

=head1 DESCRIPTION

Init video data

=head1 AUTHOR

L<sv99@inbox.ru>

=cut
