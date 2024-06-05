import sys

replacement_list_data = sys.stdin.read()
replacement_list = [line.split(' ') for line in replacement_list_data.split('\n')]

for file in sys.argv[1:]:
    with open(file, 'r') as f:
        data = f.read()
    for og, replacement in replacement_list:
        data = data.replace(og, replacement)
    with open(fie, 'w') as f:
        f.write(data)
