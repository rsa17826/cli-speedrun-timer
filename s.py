import time
import browser_cookie3

try:
  cj = browser_cookie3.brave()

  # Netscape format files require this specific header line
  print("# Netscape HTTP Cookie File")
  print("# http://curl.haxx.se/rfc/cookie_spec.html")
  print("# This is a generated file!  Do not edit.\n")

  for cookie in cj:
    if cookie.domain.endswith(".google.com") and len(cookie.name) != 32:

      # 1. Domain (Netscape format traditionally expects subdomains to start with a dot)
      # browser_cookie3 usually preserves this, but we can ensure consistency
      domain = cookie.domain

      # 2. Flag (True if domain starts with a dot, meaning it matches subdomains)
      flag = "TRUE" if domain.startswith(".") else "FALSE"

      # 3. Path
      path = cookie.path or "/"

      # 4. Secure flag
      secure = "TRUE" if cookie.secure else "FALSE"

      # 5. Expiration (Default to 0 if None/Session cookie)
      expiry = str(cookie.expires) if cookie.expires is not None else "0"

      # 6 & 7. Name and Value
      name = cookie.name
      value = cookie.value

      # Print tab-separated values
      print(f"{domain}\t{flag}\t{path}\t{secure}\t{expiry}\t{name}\t{value}")

except Exception as e:
  import sys
  print(f"# Error reading cookies: {e}", file=sys.stderr)