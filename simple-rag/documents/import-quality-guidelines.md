# Data Import Quality Guidelines

## Overview

When importing contact records from external systems, data quality issues are common. This document outlines how to identify and handle "dirty records" - contact information that contains errors, inconsistencies, or placeholder data.

## Common Sources of Dirty Data

### External System Imports

**CRM exports often contain:**

- Duplicate records with slight variations
- Incomplete records with missing required fields
- Test accounts that weren't cleaned before export
- Outdated information that wasn't maintained
- Formatting inconsistencies between systems

### Manual Data Entry

**Human entry errors include:**

- Typos and misspellings
- Inconsistent formatting
- Copy/paste errors
- Placeholder text left in fields
- Mixed character encodings

### Legacy System Migrations

**Old systems may have:**

- Non-standard data formats
- Deprecated field values
- Character encoding issues (é vs Ã©)
- Truncated fields due to length limits
- Merged or concatenated fields

## Data Quality Indicators

### High-Risk Indicators (Automatic Rejection)

These patterns strongly suggest dirty or test data:

1. All fields are identical (e.g., "test" in every field)
2. Sequential or repetitive patterns (111-1111, abc@abc.com)
3. Obvious test patterns ("test", "sample", "demo", "example")
4. Multiple records with same email but different names
5. Invalid email domains (@test.com, @example.com)
6. Suspicious phone patterns (all zeros, all nines)

### Medium-Risk Indicators (Review Required)

These patterns may indicate problems:

1. Very short company names (< 3 characters)
2. Single-word personal names
3. Generic company names ("Company", "Business")
4. Email addresses with numbers that look like versions (john1@, john2@)
5. Phone numbers with country code mismatches
6. Names with unusual capitalization

### Low-Risk Indicators (Flag for Monitoring)

Worth noting but may be legitimate:

1. International characters in names
2. Hyphenated last names
3. Single-letter middle names
4. PO Box addresses
5. Generic email domains (gmail, yahoo, hotmail)

## Data Deduplication Rules

### Exact Duplicates

**Remove if:**

- All fields match exactly (case-insensitive)
- Email addresses match (primary key)

### Near Duplicates

**Review if:**

- Email matches but name differs
- Name and company match but email differs
- Phone matches but other fields differ
- Similar names with edit distance < 3

### Merge Candidates

**Consider merging when:**

- Same email with different metadata
- Same person at different companies (job change)
- Same company with different contact persons
- Updated vs outdated versions of same record

## Field-Specific Quality Rules

### Email Field

- Required field (cannot be empty)
- Must pass format validation
- Must have valid domain
- Check against disposable email provider list
- Verify domain has MX records (if possible)

### Phone Field

- Optional but recommended
- Standardize to E.164 format if possible
- Remove decorative formatting for storage
- Validate against country code
- Flag toll-free numbers separately

### Name Fields

- First name: Required, 2-50 characters
- Last name: Required, 2-50 characters
- Avoid accepting single-word entries
- Reject entries with numbers
- Apply title case normalization

### Company Field

- Recommended field
- Standardize legal suffixes (Inc, LLC, Ltd)
- Remove website URLs
- Check against known test company names
- Flag very generic names for review

## Confidence Scoring

### Calculate Quality Score (0-100)

**Award points for:**

- Valid email format (+20)
- Email domain has MX record (+10)
- Valid phone number format (+15)
- Phone matches expected country (+5)
- Both first and last name present (+15)
- Proper name capitalization (+5)
- Company name present (+10)
- Company not in generic list (+5)
- No test patterns detected (+15)

**Deduct points for:**

- Any high-risk indicator (-50)
- Each medium-risk indicator (-15)
- Missing recommended fields (-10 each)
- Each low-risk indicator (-5)

### Quality Tiers

- **90-100**: Excellent - Auto-approve import
- **70-89**: Good - Import with standard monitoring
- **50-69**: Fair - Import but flag for review
- **30-49**: Poor - Review before import
- **0-29**: Very Poor - Reject or require manual verification

## Automated Cleaning Actions

### Safe Auto-Fixes

These changes can be applied automatically:

1. Trim leading/trailing whitespace
2. Remove multiple consecutive spaces
3. Standardize email to lowercase
4. Apply title case to names
5. Remove non-numeric characters from phone
6. Collapse line breaks in address fields

### Require Approval

These changes need human review:

1. Correcting suspected typos
2. Merging duplicate records
3. Filling in missing required fields
4. Changing data formats significantly

## Import Workflow Recommendations

### Pre-Import Phase

1. Scan file for encoding issues
2. Count total records
3. Identify column mappings
4. Preview first 10 records
5. Run quality check on sample

### Import Phase

1. Process records in batches
2. Calculate quality score per record
3. Apply safe auto-fixes
4. Log all changes made
5. Track rejection reasons

### Post-Import Phase

1. Generate summary report
2. List all flagged records
3. Identify patterns in rejections
4. Create review queue for medium-risk records
5. Update quality rules based on findings
