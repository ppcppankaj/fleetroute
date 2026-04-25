import os
import re

for root, dirs, files in os.walk('.'):
    if 'node_modules' in root or '.git' in root:
        continue
    for file in files:
        if file.endswith('.go'):
            path = os.path.join(root, file)
            with open(path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            orig = content
            
            # Fix gin 500
            content = re.sub(r'c\.JSON\((http\.StatusInternalServerError|500),\s*gin\.H\{"error":\s*err\.Error\(\)\}\)', 
                             r'c.JSON(\1, gin.H{"error": "internal server error"})', content)
            
            # Fix writeError 500
            content = re.sub(r'writeError\(w,\s*http\.StatusInternalServerError,\s*err\.Error\(\)\)',
                             r'writeError(w, http.StatusInternalServerError, "internal server error")', content)
                             
            # Fix respondError 500
            content = re.sub(r'respondError\(c,\s*http\.StatusInternalServerError,\s*err\.Error\(\)\)',
                             r'respondError(c, http.StatusInternalServerError, "internal server error")', content)
            
            if orig != content:
                print('Fixed', path)
                with open(path, 'w', encoding='utf-8') as f:
                    f.write(content)
