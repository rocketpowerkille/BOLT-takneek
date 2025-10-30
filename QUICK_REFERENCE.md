# DBCLI Quick Reference

## 🚀 How to Use DBCLI in 3 Steps

### Step 1: Run the Tool
```bash
dbcli.exe
```

### Step 2: Enter Database Information
**Source Database → Destination Database**
- Database type (mysql/oracle)
- Host, port, username, password, database name

### Step 3: Configure Migration
- Select source table and columns
- Select destination table and columns  
- Map columns (source → destination)
- Choose dump options
- Confirm and execute

## ⌨️ Keyboard Controls

| Key | Action |
|-----|--------|
| `↑↓` | Navigate lists |
| `Enter` | Select/Confirm |
| `Space` | Toggle selection |
| `h` | Show help |
| `b` | Go back |
| `q` | Quit |
| `Ctrl+C` | Emergency exit |

## 📋 Migration Process Flow

```
1. Source DB Credentials
   ↓
2. Destination DB Credentials  
   ↓
3. Source Table Selection
   ↓
4. Source Column Selection
   ↓
5. Destination Table Selection
   ↓
6. Destination Column Selection
   ↓
7. Column Mapping
   ↓
8. Dump Options
   ↓
9. Confirmation & Execution
```

## 🔧 Common Examples

### MySQL to MySQL
```
Source: mysql://localhost:3306/source_db
Destination: mysql://localhost:3306/dest_db
```

### Oracle to MySQL
```
Source: oracle://localhost:1521/XE
Destination: mysql://localhost:3306/analytics
```

### Column Mapping Example
```
Source Table: users
- id → user_id
- name → full_name
- email → email_address
- created_at → date_created

Destination Table: user_archive
```

## 📁 Files Created

| File/Folder | Purpose |
|-------------|---------|
| `config.yaml` | Configuration settings |
| `logs/dbcli.log` | Application logs |
| `checkpoints/` | Migration state files |
| `dumps/` | SQL dump files |

## 🆘 Quick Troubleshooting

| Problem | Solution |
|---------|----------|
| Connection failed | Check credentials, network, firewall |
| Permission denied | Verify database user permissions |
| Migration slow | Increase batch size in config.yaml |
| Data type error | Use validation feature, check mappings |

## 📞 Need Help?

- Press `h` during any step for context help
- Check `logs/dbcli.log` for error details
- Review `config.yaml` for settings
- See [USER_GUIDE.md](USER_GUIDE.md) for full documentation

---
**Ready to migrate?** Just run `dbcli.exe` and follow the prompts! 🎯
