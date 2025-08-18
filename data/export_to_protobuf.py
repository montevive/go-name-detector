#!/usr/bin/env python3
"""
Export pickled name data to Protocol Buffer format for Go consumption.
"""

import gzip
import pickle
import sys
from pathlib import Path

# Add the parent directory to import the generated protobuf classes
sys.path.append(str(Path(__file__).parent.parent))

def export_to_protobuf():
    """Export pickle data to protobuf format."""
    
    # Import the generated protobuf classes
    try:
        from proto import names_pb2
    except ImportError:
        print("Error: protobuf classes not found. Please run protoc to generate Python classes first.")
        return False
    
    # Load pickle files
    base_path = Path(__file__).parent.parent.parent
    first_path = base_path / "names_dataset/v3/first_names.pkl.gz"
    last_path = base_path / "names_dataset/v3/last_names.pkl.gz"
    
    print("Loading pickle files...")
    
    with gzip.open(first_path, 'rb') as f:
        first_names = pickle.load(f)
    print(f"Loaded {len(first_names):,} first names")
    
    with gzip.open(last_path, 'rb') as f:
        last_names = pickle.load(f)
    print(f"Loaded {len(last_names):,} last names")
    
    # Create protobuf datasets
    print("Converting to protobuf format...")
    
    # Convert first names
    first_dataset = names_pb2.NameDataset()
    for name, data in first_names.items():
        entry = first_dataset.entries.add()
        entry.name = name
        
        # Country probabilities
        for country, prob in data.get('country', {}).items():
            entry.country[country] = prob
            
        # Gender probabilities
        for gender, prob in data.get('gender', {}).items():
            entry.gender[gender] = prob
            
        # Ranks
        for country, rank in data.get('rank', {}).items():
            if rank is not None:
                entry.rank[country] = rank
    
    # Convert last names
    last_dataset = names_pb2.NameDataset()
    for name, data in last_names.items():
        entry = last_dataset.entries.add()
        entry.name = name
        
        # Country probabilities
        for country, prob in data.get('country', {}).items():
            entry.country[country] = prob
            
        # Gender probabilities (should be empty for last names)
        for gender, prob in data.get('gender', {}).items():
            entry.gender[gender] = prob
            
        # Ranks
        for country, rank in data.get('rank', {}).items():
            if rank is not None:
                entry.rank[country] = rank
    
    # Create combined dataset
    combined = names_pb2.CombinedNameDataset()
    combined.first_names.CopyFrom(first_dataset)
    combined.last_names.CopyFrom(last_dataset)
    
    # Write to files
    output_dir = Path(__file__).parent
    
    print("Writing protobuf files...")
    
    # Write individual files
    with open(output_dir / "first_names.pb", "wb") as f:
        f.write(first_dataset.SerializeToString())
    
    with open(output_dir / "last_names.pb", "wb") as f:
        f.write(last_dataset.SerializeToString())
    
    # Write combined file
    with open(output_dir / "combined_names.pb", "wb") as f:
        f.write(combined.SerializeToString())
    
    # Compress files
    print("Compressing files...")
    for filename in ["first_names.pb", "last_names.pb", "combined_names.pb"]:
        with open(output_dir / filename, "rb") as f_in:
            with gzip.open(output_dir / f"{filename}.gz", "wb") as f_out:
                f_out.write(f_in.read())
        # Remove uncompressed version
        (output_dir / filename).unlink()
    
    # Print file sizes
    print("\nGenerated files:")
    for filename in ["first_names.pb.gz", "last_names.pb.gz", "combined_names.pb.gz"]:
        filepath = output_dir / filename
        size_mb = filepath.stat().st_size / (1024 * 1024)
        print(f"  {filename}: {size_mb:.2f} MB")
    
    print("Export complete!")
    return True

if __name__ == "__main__":
    export_to_protobuf()