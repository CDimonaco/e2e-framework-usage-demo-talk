#!/usr/bin/env bats

@test "reject because pod has a palindrome label key not esplicitely allowed" {
  run kwctl run annotated-policy.wasm -r ./e2e/fixtures/palindrome-label-pod.json --settings-json '{"allowed_palindromes": []}'

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : ".*pod label with key level not allowed, the word is a palindrome.*") -ne 0 ]
}

@test "accept because pod has not a palindrome key label" {
  run kwctl run annotated-policy.wasm -r ./e2e/fixtures/non-palindrome-label-pod.json --settings-json '{"allowed_palindromes": []}'
  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request accepted
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

@test "accept because pod has a palindrome key label, but the label is explicitely allowed in the settings" {
  run kwctl run annotated-policy.wasm -r ./e2e/fixtures/palindrome-label-pod.json --settings-json '{"allowed_palindromes": ["level"]}'
  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request accepted
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}