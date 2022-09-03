to main


factorial:
  $1 = $1 * $0

  $0 = $0 - 1
  to factorial if $0 > 1
  to end


main:
  $0 = 12
  $1 = 1
  to factorial


end:
  write $1
