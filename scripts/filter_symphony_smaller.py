#!/usr/bin/env python3
"""
Filter out log lines where symphony message size is smaller than protobuf.

This script removes entries where symphony outperforms protobuf, keeping only
entries where protobuf is smaller or equal to symphony.
"""

import json
import argparse
import sys
from pathlib import Path


def filter_log_file(input_path: Path, output_path: Path, dry_run: bool = False) -> tuple[int, int]:
    """
    Filter a JSONL log file, removing lines where symphony < protobuf.
    
    Returns: (original_count, kept_count)
    """
    original_count = 0
    kept_lines = []
    removed_count = 0
    
    with open(input_path, 'r') as f:
        for line in f:
            original_count += 1
            line = line.strip()
            if not line:
                continue
            
            try:
                msg = json.loads(line)
                sizes = msg.get('sizes', {})
                protobuf_size = sizes.get('protobuf', 0)
                symphony_size = sizes.get('symphony', 0)
                
                # Keep line if symphony >= protobuf (i.e., protobuf wins or tie)
                if symphony_size >= protobuf_size:
                    kept_lines.append(line)
                else:
                    removed_count += 1
                    
            except json.JSONDecodeError:
                # Keep malformed lines as-is
                kept_lines.append(line)
    
    kept_count = len(kept_lines)
    
    if not dry_run:
        with open(output_path, 'w') as f:
            for line in kept_lines:
                f.write(line + '\n')
    
    return original_count, kept_count, removed_count


def main():
    parser = argparse.ArgumentParser(
        description='Filter out log lines where symphony message size is smaller than protobuf',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s                          # Filter all logs in-place
  %(prog)s --log-dir /path/to/logs  # Specify log directory
  %(prog)s --dry-run                # Preview what would be removed
  %(prog)s --output-suffix .filtered  # Create new files with suffix
        """
    )
    parser.add_argument('--log-dir', type=Path, default=Path('logs'),
                        help='Directory containing JSONL log files (default: logs)')
    parser.add_argument('--dry-run', action='store_true',
                        help='Preview changes without modifying files')
    parser.add_argument('--output-suffix', type=str, default='',
                        help='Suffix for output files (empty = in-place modification)')
    parser.add_argument('--file', type=Path,
                        help='Process a single file instead of all files in log-dir')
    
    args = parser.parse_args()
    
    # Determine files to process
    if args.file:
        if not args.file.exists():
            print(f"Error: File not found: {args.file}")
            sys.exit(1)
        log_files = [args.file]
    else:
        # Resolve log directory
        log_dir = args.log_dir
        if not log_dir.is_absolute():
            script_dir = Path(__file__).parent.parent
            log_dir = script_dir / args.log_dir
        
        if not log_dir.exists():
            print(f"Error: Log directory not found: {log_dir}")
            sys.exit(1)
        
        log_files = sorted(log_dir.glob("*.jsonl"))
    
    if not log_files:
        print("No .jsonl files found to process")
        sys.exit(1)
    
    print("=" * 80)
    print("FILTERING LOG FILES - Removing lines where symphony < protobuf")
    print("=" * 80)
    
    if args.dry_run:
        print("** DRY RUN MODE - No files will be modified **\n")
    
    total_original = 0
    total_kept = 0
    total_removed = 0
    
    for log_file in log_files:
        if args.output_suffix:
            output_file = log_file.with_suffix(args.output_suffix + log_file.suffix)
        else:
            output_file = log_file
        
        original, kept, removed = filter_log_file(log_file, output_file, args.dry_run)
        total_original += original
        total_kept += kept
        total_removed += removed
        
        pct_removed = (removed / original * 100) if original > 0 else 0
        status = "would remove" if args.dry_run else "removed"
        
        print(f"{log_file.name}")
        print(f"  Original: {original:,} lines")
        print(f"  {status.capitalize()}: {removed:,} lines ({pct_removed:.1f}%)")
        print(f"  Kept: {kept:,} lines")
        if args.output_suffix and not args.dry_run:
            print(f"  Output: {output_file.name}")
        print()
    
    print("=" * 80)
    print("SUMMARY")
    print("=" * 80)
    total_pct_removed = (total_removed / total_original * 100) if total_original > 0 else 0
    print(f"Total original lines: {total_original:,}")
    print(f"Total removed:        {total_removed:,} ({total_pct_removed:.1f}%)")
    print(f"Total kept:           {total_kept:,}")
    
    if args.dry_run:
        print("\n** This was a dry run. Run without --dry-run to apply changes. **")


if __name__ == '__main__':
    main()

