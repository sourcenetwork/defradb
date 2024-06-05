import sys

replacement_list_data = sys.stdin.read()
replacement_list = [line.split(' ') for line in replacement_list_data.split('\n') if line]
print(replacement_list)

for file in sys.argv[1:]:
    with open(file, 'r') as f:
        data = f.read()
    for entry in replacement_list:
        og, repl = entry[0], entry[1]
        data = data.replace(og, repl)
    with open(file, 'w') as f:
        f.write(data)
