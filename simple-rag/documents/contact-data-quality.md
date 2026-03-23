# Contact Data Quality Standards

## Email Address Validation

### Valid Email Format

A valid email address must follow these criteria:

- Contains exactly one @ symbol
- Has a local part before the @ (cannot be empty)
- Has a domain part after the @ (cannot be empty)
- Domain must contain at least one dot (.)
- No spaces allowed anywhere in the email
- Local part can contain letters, numbers, dots, hyphens, and underscores
- Domain part should be a valid domain name

### Common Email Issues

**Invalid formats that indicate dirty data:**

- Missing @ symbol: "john.doegmail.com"
- Multiple @ symbols: "john@@company.com"
- Spaces in email: "john doe@company.com"
- Missing domain: "john@"
- Missing local part: "@company.com"
- Invalid characters: "john$%^@company.com"
- Incomplete domain: "john@company" (missing TLD)

### Email Cleaning Recommendations

1. Trim whitespace from both ends
2. Convert to lowercase for consistency
3. Check for common typos: "gamil.com" → "gmail.com"
4. Reject placeholder emails: "test@test.com", "noreply@noreply.com"
5. Flag disposable email domains: mailinator.com, guerrillamail.com

## Phone Number Validation

### Valid Phone Number Formats

Phone numbers should meet these requirements:

- Contains only digits, spaces, hyphens, parentheses, and plus sign
- Has appropriate length (7-15 digits for international format)
- Includes country code for international numbers (+1, +44, etc.)
- Area code in parentheses or with hyphens is acceptable

### Acceptable Formats

- US Format: (555) 123-4567
- International: +1-555-123-4567
- Simple: 555-123-4567
- Minimal: 5551234567

### Phone Number Issues

**Invalid patterns that indicate dirty data:**

- Sequential numbers: 111-1111, 123-4567
- All zeros: 000-0000
- Clearly fake: 555-0000, 999-9999
- Letters mixed in: 555-CALL
- Too few digits: 123-45
- Too many digits: 555-123-4567890
- Special characters: 555@123#4567

### Phone Cleaning Recommendations

1. Remove all non-numeric characters except + at start
2. Validate length (minimum 7 digits)
3. Check for sequential or repeated patterns
4. Verify area code is valid
5. Cross-reference with country code if present

## Name Validation

### Valid Name Format

Person names should follow these guidelines:

- Contains at least first and last name
- Uses proper capitalization
- Contains only letters, spaces, hyphens, and apostrophes
- No numbers or special characters (except hyphen and apostrophe)
- Reasonable length (2-50 characters for each part)

### Name Quality Issues

**Problems that indicate dirty data:**

- Single word names: "John" (missing last name)
- ALL CAPS: "JOHN SMITH"
- all lowercase: "john smith"
- Numbers in name: "John Smith 123"
- Special characters: "John@Smith"
- Test names: "Test User", "First Last", "Name Here"
- Too short: "A B"
- Too long: Names over 100 characters
- Multiple consecutive spaces
- Leading/trailing spaces

### Name Cleaning Recommendations

1. Trim leading and trailing whitespace
2. Capitalize first letter of each word (title case)
3. Remove numbers and special characters
4. Collapse multiple spaces to single space
5. Reject obvious test names
6. Ensure both first and last name are present
7. Flag single-character names for review

## Company Name Validation

### Valid Company Format

Company names should meet these criteria:

- Not empty or null
- Minimum 2 characters
- Can contain letters, numbers, spaces, and common punctuation
- Proper capitalization
- No HTML or code snippets

### Company Name Issues

**Problems indicating dirty data:**

- Empty or null company name
- Generic placeholders: "Company", "N/A", "None"
- Personal names instead of company: "John Smith"
- All caps: "ABC CORPORATION"
- URL instead of name: "www.company.com"
- HTML tags: "<company name>"
- Executive titles: "CEO John Smith"

### Company Cleaning Recommendations

1. Remove URLs and convert to proper company names
2. Remove executive titles
3. Standardize legal entity suffixes: Inc., LLC, Ltd.
4. Remove extraneous punctuation
5. Apply proper title casing
6. Flag generic placeholders for manual review
