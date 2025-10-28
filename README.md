# Web Watcher

**Web Watcher** is a lightweight **Web Application Firewall (WAF)** written in **Go**.  
It provides protection against common web security threats such as **SQL Injection**, **Cross-Site Scripting (XSS)**, and **brute-force or abusive requests** using **rate limiting** and **captcha-based verification**.

---

## Features

### SQL Injection Detection
- Automatically scans and blocks requests containing suspicious SQL patterns.
- Prevents attacks such as:
  - `' OR '1'='1`
  - `UNION SELECT`
  - `--` and other SQL comment-based exploits.

### Cross-Site Scripting (XSS) Detection
- Detects malicious scripts embedded in request parameters.
- Neutralizes payloads like:
  '''
  <script>alert(1)</script>
  '''
### Rate Limiting 
- Limits the number of requests per IP within a time span.
- Helps mitigate: 
  - Brute-force login attempts.
  - DDoS-like request flood.
### Captcha Verification
- Integrates interactive slider captcha before allowing login or sensitive requests.
- Prevent bots from bypassing authentication.
