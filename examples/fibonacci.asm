to main


fibonacci:
  to end if $0 < 2

  $3 = $1 + $2
  $0 = $0 - 1
  to end if $0 < 1

  $1 = $2
  $2 = $3
  to fibonacci


main:
  $0 = 56
  $1 = 1
  $2 = 1
  $3 = 1
  to fibonacci


end:
  print $1
