#!/usr/bin/env python3
"""
Analyze serialization message sizes across different formats.

This script analyzes JSONL log files to understand why symphony can outperform
protobuf for certain message types and under what conditions.
"""

import json
import os
import sys
import argparse
from pathlib import Path
from collections import defaultdict
from dataclasses import dataclass, field
from typing import Dict, List, Any, Optional
import statistics


@dataclass
class SizeStats:
    """Statistics for a serialization format's message sizes."""
    sizes: List[int] = field(default_factory=list)
    
    def add(self, size: int):
        self.sizes.append(size)
    
    @property
    def count(self) -> int:
        return len(self.sizes)
    
    @property
    def total(self) -> int:
        return sum(self.sizes)
    
    @property
    def mean(self) -> float:
        return statistics.mean(self.sizes) if self.sizes else 0
    
    @property
    def median(self) -> float:
        return statistics.median(self.sizes) if self.sizes else 0
    
    @property
    def min(self) -> int:
        return min(self.sizes) if self.sizes else 0
    
    @property
    def max(self) -> int:
        return max(self.sizes) if self.sizes else 0
    
    @property
    def stdev(self) -> float:
        return statistics.stdev(self.sizes) if len(self.sizes) > 1 else 0


@dataclass
class MessageTypeStats:
    """Statistics for a specific message type."""
    message_type: str
    direction: str
    formats: Dict[str, SizeStats] = field(default_factory=lambda: defaultdict(SizeStats))
    payloads: List[Dict] = field(default_factory=list)
    
    def add_message(self, sizes: Dict[str, int], payload: Dict):
        for fmt, size in sizes.items():
            self.formats[fmt].add(size)
        # Store sample payloads for analysis (limit to avoid memory issues)
        if len(self.payloads) < 100:
            self.payloads.append(payload)


def analyze_payload_structure(payload: Dict, depth: int = 0) -> Dict[str, Any]:
    """Analyze the structure of a payload to understand what affects serialization."""
    analysis = {
        "type": type(payload).__name__,
        "size": 0,
        "string_bytes": 0,
        "num_fields": 0,
        "array_lengths": [],
        "nested_objects": 0,
    }
    
    if isinstance(payload, dict):
        analysis["num_fields"] = len(payload)
        for key, value in payload.items():
            if isinstance(value, str):
                analysis["string_bytes"] += len(value.encode('utf-8'))
            elif isinstance(value, list):
                analysis["array_lengths"].append(len(value))
                for item in value:
                    if isinstance(item, dict):
                        analysis["nested_objects"] += 1
                        sub = analyze_payload_structure(item, depth + 1)
                        analysis["string_bytes"] += sub["string_bytes"]
                        analysis["nested_objects"] += sub["nested_objects"]
            elif isinstance(value, dict):
                analysis["nested_objects"] += 1
                sub = analyze_payload_structure(value, depth + 1)
                analysis["string_bytes"] += sub["string_bytes"]
                analysis["nested_objects"] += sub["nested_objects"]
    
    return analysis


def load_log_files(log_dir: Path) -> List[Dict]:
    """Load all JSONL log files from a directory."""
    messages = []
    log_files = list(log_dir.glob("*.jsonl"))
    
    if not log_files:
        print(f"No .jsonl files found in {log_dir}")
        return messages
    
    for log_file in sorted(log_files):
        print(f"Loading {log_file.name}...")
        with open(log_file, 'r') as f:
            for line_num, line in enumerate(f, 1):
                try:
                    msg = json.loads(line.strip())
                    if 'sizes' in msg and 'message_type' in msg:
                        msg['_source_file'] = log_file.name
                        messages.append(msg)
                except json.JSONDecodeError as e:
                    print(f"  Warning: Invalid JSON at line {line_num}: {e}")
    
    return messages


def compute_statistics(messages: List[Dict]) -> Dict[str, MessageTypeStats]:
    """Compute statistics for each message type."""
    stats = {}
    
    for msg in messages:
        msg_type = msg['message_type']
        direction = msg.get('direction', 'unknown')
        key = f"{msg_type}_{direction}"
        
        if key not in stats:
            stats[key] = MessageTypeStats(msg_type, direction)
        
        stats[key].add_message(msg['sizes'], msg.get('payload', {}))
    
    return stats


def format_size(size: float) -> str:
    """Format a size in bytes with appropriate units."""
    if size >= 1024:
        return f"{size/1024:.2f} KB"
    return f"{size:.1f} B"


def print_comparison_table(stats: Dict[str, MessageTypeStats], formats: List[str]):
    """Print a comparison table of message sizes across formats."""
    print("\n" + "=" * 120)
    print("MESSAGE SIZE COMPARISON BY TYPE")
    print("=" * 120)
    
    # Header
    header = f"{'Message Type':<40} {'Dir':<8} {'Count':>8}"
    for fmt in formats:
        header += f" {fmt:>12}"
    print(header)
    print("-" * 120)
    
    # Sort by message type
    for key in sorted(stats.keys()):
        stat = stats[key]
        row = f"{stat.message_type:<40} {stat.direction:<8} {stat.formats[formats[0]].count:>8}"
        for fmt in formats:
            row += f" {stat.formats[fmt].mean:>12.1f}"
        print(row)
    
    print("-" * 120)


def print_symphony_vs_protobuf_analysis(stats: Dict[str, MessageTypeStats]):
    """Analyze when symphony outperforms protobuf and why."""
    print("\n" + "=" * 120)
    print("SYMPHONY vs PROTOBUF DETAILED ANALYSIS")
    print("=" * 120)
    
    symphony_wins = []
    protobuf_wins = []
    ties = []
    
    for key, stat in stats.items():
        pb_stats = stat.formats.get('protobuf')
        sym_stats = stat.formats.get('symphony')
        
        if not pb_stats or not sym_stats or pb_stats.count == 0:
            continue
        
        pb_mean = pb_stats.mean
        sym_mean = sym_stats.mean
        
        if pb_mean == 0 and sym_mean == 0:
            ties.append((key, stat, 0, 0, 0))
        elif sym_mean < pb_mean:
            savings = pb_mean - sym_mean
            savings_pct = (savings / pb_mean * 100) if pb_mean > 0 else 0
            symphony_wins.append((key, stat, pb_mean, sym_mean, savings_pct))
        elif pb_mean < sym_mean:
            overhead = sym_mean - pb_mean
            overhead_pct = (overhead / pb_mean * 100) if pb_mean > 0 else 0
            protobuf_wins.append((key, stat, pb_mean, sym_mean, overhead_pct))
        else:
            ties.append((key, stat, pb_mean, sym_mean, 0))
    
    # Symphony wins
    print(f"\n🎯 SYMPHONY WINS ({len(symphony_wins)} message types):")
    print("-" * 100)
    if symphony_wins:
        symphony_wins.sort(key=lambda x: x[4], reverse=True)  # Sort by savings %
        print(f"{'Message Type':<50} {'PB Mean':>10} {'SYM Mean':>10} {'Savings':>10} {'%':>8}")
        print("-" * 100)
        for key, stat, pb_mean, sym_mean, savings_pct in symphony_wins:
            print(f"{stat.message_type} ({stat.direction})"[:50].ljust(50) + 
                  f" {pb_mean:>10.1f} {sym_mean:>10.1f} {pb_mean - sym_mean:>10.1f} {savings_pct:>7.1f}%")
    
    # Protobuf wins
    print(f"\n📊 PROTOBUF WINS ({len(protobuf_wins)} message types):")
    print("-" * 100)
    if protobuf_wins:
        protobuf_wins.sort(key=lambda x: x[4], reverse=True)  # Sort by overhead %
        print(f"{'Message Type':<50} {'PB Mean':>10} {'SYM Mean':>10} {'Overhead':>10} {'%':>8}")
        print("-" * 100)
        for key, stat, pb_mean, sym_mean, overhead_pct in protobuf_wins:
            print(f"{stat.message_type} ({stat.direction})"[:50].ljust(50) + 
                  f" {pb_mean:>10.1f} {sym_mean:>10.1f} {sym_mean - pb_mean:>10.1f} {overhead_pct:>7.1f}%")
    
    # Ties
    if ties:
        print(f"\n🤝 TIES ({len(ties)} message types - typically empty messages)")
    
    return symphony_wins, protobuf_wins


def analyze_why_symphony_wins(stats: Dict[str, MessageTypeStats], symphony_wins: List):
    """Deep dive into why symphony wins for certain message types."""
    print("\n" + "=" * 120)
    print("WHY SYMPHONY OUTPERFORMS PROTOBUF - STRUCTURAL ANALYSIS")
    print("=" * 120)
    
    for key, stat, pb_mean, sym_mean, savings_pct in symphony_wins[:10]:  # Top 10
        print(f"\n📝 {stat.message_type} ({stat.direction})")
        print(f"   Symphony saves {pb_mean - sym_mean:.1f} bytes ({savings_pct:.1f}%) per message")
        
        # Analyze payload structure
        if stat.payloads:
            # Get a representative sample
            sample = stat.payloads[0] if stat.payloads else {}
            analysis = analyze_payload_structure(sample)
            
            print(f"   Payload structure:")
            print(f"     - Fields: {analysis['num_fields']}")
            print(f"     - String bytes: {analysis['string_bytes']}")
            print(f"     - Array lengths: {analysis['array_lengths']}")
            print(f"     - Nested objects: {analysis['nested_objects']}")
            
            # Hypothesis based on structure
            if analysis['num_fields'] <= 3 and not analysis['array_lengths']:
                print(f"   → Simple flat structure: Symphony's minimal framing overhead wins")
            elif analysis['string_bytes'] == 0 and analysis['num_fields'] > 0:
                print(f"   → Numeric-only fields: Symphony's efficient number encoding wins")
            elif analysis['array_lengths'] and max(analysis['array_lengths']) <= 2:
                print(f"   → Small arrays: Symphony's compact array encoding is efficient")


def analyze_why_protobuf_wins(stats: Dict[str, MessageTypeStats], protobuf_wins: List):
    """Deep dive into why protobuf wins for certain message types."""
    print("\n" + "=" * 120)
    print("WHY PROTOBUF OUTPERFORMS SYMPHONY - STRUCTURAL ANALYSIS")
    print("=" * 120)
    
    for key, stat, pb_mean, sym_mean, overhead_pct in protobuf_wins[:10]:  # Top 10
        print(f"\n📝 {stat.message_type} ({stat.direction})")
        print(f"   Symphony adds {sym_mean - pb_mean:.1f} bytes ({overhead_pct:.1f}%) overhead per message")
        
        # Analyze payload structure
        if stat.payloads:
            sample = stat.payloads[0] if stat.payloads else {}
            analysis = analyze_payload_structure(sample)
            
            print(f"   Payload structure:")
            print(f"     - Fields: {analysis['num_fields']}")
            print(f"     - String bytes: {analysis['string_bytes']}")
            print(f"     - Array lengths: {analysis['array_lengths']}")
            print(f"     - Nested objects: {analysis['nested_objects']}")
            
            # Hypothesis based on structure
            if analysis['string_bytes'] > 100:
                print(f"   → Large string content: Protobuf's length-prefixed strings are efficient")
            if analysis['nested_objects'] > 3:
                print(f"   → Deep nesting: Protobuf's schema-driven encoding wins")
            if analysis['array_lengths'] and max(analysis['array_lengths']) > 3:
                print(f"   → Large arrays: Protobuf's repeated field encoding is compact")


def print_size_distribution(stats: Dict[str, MessageTypeStats]):
    """Print size distribution analysis."""
    print("\n" + "=" * 120)
    print("SIZE DISTRIBUTION BY MESSAGE TYPE")
    print("=" * 120)
    
    formats = ['protobuf', 'symphony', 'symphony_hybrid', 'flatbuffers', 'capnproto']
    
    for key in sorted(stats.keys()):
        stat = stats[key]
        print(f"\n{stat.message_type} ({stat.direction}) - {stat.formats['protobuf'].count} messages")
        print("-" * 80)
        print(f"{'Format':<18} {'Mean':>10} {'Median':>10} {'Min':>10} {'Max':>10} {'StdDev':>10}")
        print("-" * 80)
        
        for fmt in formats:
            if fmt in stat.formats:
                fs = stat.formats[fmt]
                print(f"{fmt:<18} {fs.mean:>10.1f} {fs.median:>10.1f} {fs.min:>10} {fs.max:>10} {fs.stdev:>10.1f}")


def print_aggregate_summary(stats: Dict[str, MessageTypeStats]):
    """Print aggregate summary across all message types."""
    print("\n" + "=" * 120)
    print("AGGREGATE SUMMARY")
    print("=" * 120)
    
    formats = ['protobuf', 'symphony', 'symphony_hybrid', 'flatbuffers', 'capnproto']
    totals = {fmt: 0 for fmt in formats}
    counts = {fmt: 0 for fmt in formats}
    
    for stat in stats.values():
        for fmt in formats:
            if fmt in stat.formats:
                totals[fmt] += stat.formats[fmt].total
                counts[fmt] += stat.formats[fmt].count
    
    print(f"\n{'Format':<18} {'Total Bytes':>15} {'Msg Count':>12} {'Avg Size':>12} {'vs Protobuf':>12}")
    print("-" * 80)
    
    pb_total = totals.get('protobuf', 1)
    for fmt in formats:
        if counts[fmt] > 0:
            avg = totals[fmt] / counts[fmt]
            ratio = (totals[fmt] / pb_total * 100) if pb_total > 0 else 0
            print(f"{fmt:<18} {totals[fmt]:>15,} {counts[fmt]:>12,} {avg:>12.1f} {ratio:>11.1f}%")


def print_symphony_hybrid_comparison(stats: Dict[str, MessageTypeStats]):
    """Compare symphony vs symphony_hybrid."""
    print("\n" + "=" * 120)
    print("SYMPHONY vs SYMPHONY_HYBRID COMPARISON")
    print("=" * 120)
    
    print(f"\n{'Message Type':<45} {'Direction':<10} {'Symphony':>10} {'Hybrid':>10} {'Diff':>10}")
    print("-" * 95)
    
    for key in sorted(stats.keys()):
        stat = stats[key]
        sym = stat.formats.get('symphony')
        hyb = stat.formats.get('symphony_hybrid')
        
        if sym and hyb and sym.count > 0:
            diff = hyb.mean - sym.mean
            print(f"{stat.message_type:<45} {stat.direction:<10} {sym.mean:>10.1f} {hyb.mean:>10.1f} {diff:>+10.1f}")


def main():
    parser = argparse.ArgumentParser(
        description='Analyze serialization message sizes across different formats',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s                          # Analyze logs in default location
  %(prog)s --log-dir /path/to/logs  # Specify log directory
  %(prog)s --detailed               # Show detailed per-message-type stats
  %(prog)s --export results.json    # Export results to JSON
        """
    )
    parser.add_argument('--log-dir', type=Path, default=Path('logs'),
                        help='Directory containing JSONL log files (default: logs)')
    parser.add_argument('--detailed', action='store_true',
                        help='Show detailed size distribution for each message type')
    parser.add_argument('--export', type=Path,
                        help='Export analysis results to JSON file')
    
    args = parser.parse_args()
    
    # Resolve log directory
    log_dir = args.log_dir
    if not log_dir.is_absolute():
        # Try relative to script location first
        script_dir = Path(__file__).parent.parent
        log_dir = script_dir / args.log_dir
    
    if not log_dir.exists():
        print(f"Error: Log directory not found: {log_dir}")
        sys.exit(1)
    
    print(f"Analyzing logs from: {log_dir}")
    print("=" * 120)
    
    # Load and analyze
    messages = load_log_files(log_dir)
    if not messages:
        print("No messages found to analyze.")
        sys.exit(1)
    
    print(f"\nLoaded {len(messages):,} messages")
    
    stats = compute_statistics(messages)
    print(f"Found {len(stats)} unique message type/direction combinations")
    
    # Get available formats
    all_formats = set()
    for stat in stats.values():
        all_formats.update(stat.formats.keys())
    formats = sorted(all_formats)
    
    print(f"Serialization formats: {', '.join(formats)}")
    
    # Print analyses
    print_comparison_table(stats, formats)
    symphony_wins, protobuf_wins = print_symphony_vs_protobuf_analysis(stats)
    analyze_why_symphony_wins(stats, symphony_wins)
    analyze_why_protobuf_wins(stats, protobuf_wins)
    print_symphony_hybrid_comparison(stats)
    print_aggregate_summary(stats)
    
    if args.detailed:
        print_size_distribution(stats)
    
    # Export if requested
    if args.export:
        export_data = {}
        for key, stat in stats.items():
            export_data[key] = {
                'message_type': stat.message_type,
                'direction': stat.direction,
                'formats': {
                    fmt: {
                        'count': fs.count,
                        'total': fs.total,
                        'mean': fs.mean,
                        'median': fs.median,
                        'min': fs.min,
                        'max': fs.max,
                        'stdev': fs.stdev
                    }
                    for fmt, fs in stat.formats.items()
                }
            }
        
        with open(args.export, 'w') as f:
            json.dump(export_data, f, indent=2)
        print(f"\nExported results to {args.export}")
    
    print("\n" + "=" * 120)
    print("ANALYSIS COMPLETE")
    print("=" * 120)


if __name__ == '__main__':
    main()

