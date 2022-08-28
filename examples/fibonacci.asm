to main


fibonacci:
  to end if $0 < 2

  $3 = $1 + $2
  $0 = $0 - 1
  to end if $0 < 1

  $1 = $2 + 0
  $2 = $3 + 0
  to fibonacci


main:
  $0 = 56 + 0
  $1 = 1 + 0
  $2 = 1 + 0
  $3 = 1 + 0
  to fibonacci


end:
  print $1
