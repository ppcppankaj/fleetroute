import os
import re

directories_to_check = [
    'services',
    'scripts',
    'admin-panel',
]
files_to_check = [
    'docker-compose.yml',
    'docker-compose.saas.yml',
]

def replace_in_file(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
            if 'sslmode=disable' in content:
                # Do not replace if it's in the comment explaining 'sslmode=disable on ALL DATABASE_URL values below.'
                # but let's just replace all since the task says 'sslmode=disable on all 17 DATABASE_URLs. Still 18 occurrences. Need to change to sslmode=require.'
                
                # Replace 'sslmode=disable' with 'sslmode=require'
                new_content = content.replace('sslmode=disable', 'sslmode=require')
                
                with open(filepath, 'w', encoding='utf-8') as fw:
                    fw.write(new_content)
                print(f"Updated {filepath}")
    except Exception as e:
        print(f"Error processing {filepath}: {e}")

# Process directories
for d in directories_to_check:
    for root, dirs, files in os.walk(d):
        for file in files:
            if file.endswith('.go') or file.endswith('.sh') or file.endswith('.yml') or file.endswith('.yaml'):
                replace_in_file(os.path.join(root, file))

# Process individual files
for f in files_to_check:
    if os.path.exists(f):
        replace_in_file(f)

