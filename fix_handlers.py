import re

with open('api-service/internal/handler/handlers.go', 'r') as f:
    lines = f.readlines()

out = []
for i in range(len(lines)):
    line = lines[i]
    if 'rows, _ := h.pool.Query(' in line:
        line = line.replace('rows, _ :=', 'rows, err :=')
    
    if 'defer rows.Close()' in line and i > 0 and ')' in lines[i-1]:
        # we found the end of the query call
        out.append('\tif err != nil {\n\t\th.logger.Error("database error", zap.Error(err))\n\t\trespondError(c, http.StatusInternalServerError, "database error")\n\t\treturn\n\t}\n')
    
    if 'vals, _ := rows.Values()' in line:
        line = line.replace('vals, _ := rows.Values()', 'vals, err := rows.Values()\n\t\tif err != nil {\n\t\t\th.logger.Error("scan error", zap.Error(err))\n\t\t\tcontinue\n\t\t}')
        
    if '.Scan(&id)' in line and 'respondCreated(c, gin.H{"id": id})' in lines[i+1]:
        line = line + '\n\tif id == "" {\n\t\th.logger.Error("insert failed")\n\t\trespondError(c, http.StatusInternalServerError, "database error")\n\t\treturn\n\t}'

    out.append(line)

with open('api-service/internal/handler/handlers.go', 'w') as f:
    f.writelines(out)
