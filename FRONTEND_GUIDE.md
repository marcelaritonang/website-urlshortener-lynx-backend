# Frontend Integration Guide

## Quick Start

**API Base URL:** `http://localhost:8080`

## 📧 Forgot Password Flow

1. User clicks "Forgot Password"
2. Call: `POST /v1/auth/forgot-password`
3. Email sent to user's Gmail
4. User clicks link: `/reset-password?token=xxx`
5. Call: `POST /v1/auth/reset-password`
6. Redirect to login

## 📝 Example Code

```typescript
// forgot-password.tsx
const handleForgotPassword = async (email: string) => {
  const res = await fetch("http://localhost:8080/v1/auth/forgot-password", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });

  const data = await res.json();
  if (data.success) {
    alert("Check your email for reset link");
  }
};

// reset-password.tsx
const handleResetPassword = async (token: string, password: string) => {
  const res = await fetch("http://localhost:8080/v1/auth/reset-password", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token, new_password: password }),
  });

  const data = await res.json();
  if (data.success) {
    router.push("/login");
  }
};
```

## ⚠️ Important

- Token expires in 1 hour
- Password must be 8+ chars with uppercase, lowercase, number, special char
- Always check `data.success` before proceeding

## 🔗 More Details

See `API_DOCS.md` for full endpoint list (not in git, request from backend team)
