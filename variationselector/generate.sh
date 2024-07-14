#!/bin/bash
echo -e "$(
	curl -s https://www.unicode.org/Public/15.1.0/ucd/emoji/emoji-variation-sequences.txt \
	| grep FE0F \
	| awk '{ printf("\\U%8s\n", $1) }' \
	| sed 's/ /0/g'
)" | jq -RcM '[inputs]' > emojis-with-variations.json

# Why does this need a \n at the beginning to avoid eating the first emoji?!?!
echo -e "\n$(
	curl -s https://unicode.org/Public/emoji/15.1/emoji-test.txt \
	| grep '; fully-qualified' \
	| grep FE0F \
	| sed -E 's/\s+;.*//g' \
	| awk '{ for (i = 1; i <= NF; i++) {printf("\\U%8s", $i) }; printf("\n") }' \
	| sed 's/ /0/g'
)" | jq -RcM '[inputs]' > fully-qualified-variations.json

python <<EOF
import json
with open("fully-qualified-variations.json") as f:
	fully_qualified = set(json.load(f))
with open("emojis-with-variations.json") as f:
	emojis_with_variations = json.load(f)
emojis_with_variations = [x for x in emojis_with_variations if f"{x}\ufe0f" not in fully_qualified]
with open("emojis-with-extra-variations.json", "w") as f:
	json.dump(emojis_with_variations, f, ensure_ascii=False, separators=(",",":"))
EOF
rm -f emojis-with-variations.json
