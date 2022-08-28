to main


factorial:
  $1 = $1 * $0

  $0 = $0 - 1
  to factorial if $0 > 1
  to end


main:
  $0 = 12 + 0
  $1 = 1 + 0
  to factorial


end:
  print $1
