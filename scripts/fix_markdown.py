#!/usr/bin/env python3
"""
Fix markdown formatting issues reported by Codacy (markdownlint).
- MD022: Add blank lines before/after headings
- MD032: Add blank lines before/after lists
"""

import re
import sys
from pathlib import Path


def fix_markdown_file(filepath: Path) -> bool:
    """Fix markdown formatting issues in a file."""
    content = filepath.read_text(encoding='utf-8')
    original = content
    
    # MD022: Ensure blank line after headings (before non-blank content)
    # Match: heading followed immediately by non-blank line
    content = re.sub(
        r'(^#{1,6} .+$)\n([^\n#])',
        r'\1\n\n\2',
        content,
        flags=re.MULTILINE
    )
    
    # MD032: Ensure blank line before lists
    # Match: non-blank line followed by list item
    content = re.sub(
        r'([^\n])\n(^[-*+] )',
        r'\1\n\n\2',
        content,
        flags=re.MULTILINE
    )
    
    # Match: non-blank line followed by numbered list
    content = re.sub(
        r'([^\n])\n(^\d+\. )',
        r'\1\n\n\2',
        content,
        flags=re.MULTILINE
    )
    
    # MD032: Ensure blank line after lists
    # Match: list item followed by non-blank, non-list line
    content = re.sub(
        r'(^[-*+] .+$)\n([^\n-*+])',
        r'\1\n\n\2',
        content,
        flags=re.MULTILINE
    )
    
    # Match: numbered list item followed by non-blank, non-list line
    content = re.sub(
        r'(^\d+\. .+$)\n([^\n\d])',
        r'\1\n\n\2',
        content,
        flags=re.MULTILINE
    )
    
    # Remove triple+ blank lines (keep max 2)
    content = re.sub(r'\n{4,}', '\n\n\n', content)
    
    if content != original:
        filepath.write_text(content, encoding='utf-8')
        return True
    return False


def main():
    """Fix markdown files."""
    files_to_fix = [
        Path('docs/UI_IMPLEMENTATION.md'),
        Path('web/README.md'),
    ]
    
    fixed_count = 0
    for filepath in files_to_fix:
        if filepath.exists():
            if fix_markdown_file(filepath):
                print(f'✅ Fixed: {filepath}')
                fixed_count += 1
            else:
                print(f'✓ No changes needed: {filepath}')
        else:
            print(f'⚠️  Not found: {filepath}')
    
    print(f'\n📝 Fixed {fixed_count} file(s)')
    return 0


if __name__ == '__main__':
    sys.exit(main())
