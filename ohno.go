package main

const src64 = `
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
rkJggg==
`
