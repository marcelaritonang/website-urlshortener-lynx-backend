# Forgot Password Setup Guide

## Setup Gmail SMTP

1. **Enable 2-Factor Authentication** di akun Gmail Anda
2. **Generate App Password**:

   - Buka https://myaccount.google.com/apppasswords
   - Pilih "Mail" dan device Anda
   - Copy password yang dihasilkan (16 karakter)

3. **Update file `.env`**:

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-16-char-app-password
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_FROM_NAME=URL Shortener
FRONTEND_URL=http://localhost:3000
```

## API Endpoints

### 1. Request Password Reset

```bash
POST /api/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response:**

```json
{
  "message": "If the email exists, a password reset link has been sent"
}
```

### 2. Confirm Password Reset

```bash
POST /api/auth/reset-password
Content-Type: application/json

{
  "token": "reset-token-from-email",
  "new_password": "NewSecurePassword123!"
}
```

**Response:**

```json
{
  "message": "Password has been reset successfully"
}
```

## Testing

1. **Request reset**:

```bash
curl -X POST http://localhost:8080/api/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

2. **Check email** untuk link reset password

3. **Reset password**:

```bash
curl -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_TOKEN","new_password":"NewPassword123!"}'
```

## Migration

Jalankan migration untuk update database schema:

```bash
go run main.go migrate
```

## Security Notes

- Token expire dalam 1 jam
- Password harus minimal 8 karakter dengan uppercase, lowercase, number, dan special character
- Email tidak mengungkapkan apakah user exists (security best practice)
- Reset token disimpan di database dengan index untuk performa
- Token di-clear setelah digunakan

## Troubleshooting

**Email tidak terkirim:**

- Pastikan App Password benar (bukan password Gmail biasa)
- Check firewall/antivirus tidak block port 587
- Verifikasi SMTP credentials di `.env`

**Token invalid:**

- Token hanya valid 1 jam
- Token hanya bisa digunakan sekali
- Check database apakah token tersimpan dengan benar
