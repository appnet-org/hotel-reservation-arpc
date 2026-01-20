#!/usr/bin/python
#
# Generator for Hotel Reservation Application
# Based on original wrk/Lua script logic

"""
Generate curl requests for Hotel Reservation application based on wrk workload pattern.

This script generates 1000 (or custom number) of curl requests that simulate user behavior
matching the weighted task distribution from the original Lua script:
- search_hotel: 60%
- recommend: 39%
- user_login: 0.5%
- reserve: 0.5%

Usage Examples:
    # Generate 1000 curl commands and print examples
    python generate_hotel_requests.py

    # Generate and save to file
    python generate_hotel_requests.py --output hotel_load.sh

    # Execute requests against specific host
    python generate_hotel_requests.py --base-url http://localhost:11000 --execute
    python generate_curl_requests.py --base-url http://10.96.88.88:11000 --num-requests 10000 --execute

Arguments:
    --num-requests N    Number of requests to generate (default: 1000)
    --base-url URL      Base URL (default: http://localhost:11000)
    --execute           Execute the curl commands
    --output FILE       Output file to save commands
"""

import random
import subprocess
import sys
import argparse

# Base URL - adjust via arguments
BASE_URL = "http://localhost:11000"

def get_lat_lon():
    """Replicates the math.random logic for coordinates from the Lua script"""
    # 38.0235 + (math.random(0, 481) - 240.5)/1000.0
    lat = 38.0235 + (random.randint(0, 481) - 240.5) / 1000.0
    # -122.095 + (math.random(0, 325) - 157.0)/1000.0
    lon = -122.095 + (random.randint(0, 325) - 157.0) / 1000.0
    return lat, lon

def get_dates():
    """Replicates the specific April 2015 date logic from the Lua script"""
    in_date = random.randint(9, 23)
    out_date = random.randint(in_date + 1, 24)
    
    in_date_str = f"2015-04-{in_date:02d}"
    out_date_str = f"2015-04-{out_date:02d}"
    
    return in_date_str, out_date_str

def get_user_creds():
    """Replicates the Cornell_{id} user generation logic"""
    user_id = random.randint(0, 500)
    username = f"Cornell_{user_id}"
    # Password logic: repeated ID string (0 to 9 loop implies 10 times)
    password = str(user_id) * 10
    return str(user_id), username, password

def generate_search_hotel_curl():
    """Generate curl for searching hotels"""
    in_date, out_date = get_dates()
    lat, lon = get_lat_lon()
    
    url = (f"{BASE_URL}/hotels?inDate={in_date}"
           f"&outDate={out_date}&lat={lat}&lon={lon}")
    
    return f'curl -s "{url}"'

def generate_recommend_curl():
    """Generate curl for recommendations"""
    coin = random.random()
    if coin < 0.33:
        req_param = "dis"
    elif coin < 0.66:
        req_param = "rate"
    else:
        req_param = "price"

    lat, lon = get_lat_lon()
    
    url = (f"{BASE_URL}/recommendations?require={req_param}"
           f"&lat={lat}&lon={lon}")
    
    return f'curl -s "{url}"'

def generate_user_login_curl():
    """Generate curl for user login"""
    _, username, password = get_user_creds()
    
    # Note: The Lua script sends params in query string for POST
    url = f"{BASE_URL}/user?username={username}&password={password}"
    
    return f'curl -s -X POST "{url}"'

def generate_reserve_curl():
    """Generate curl for making a reservation"""
    in_date, out_date = get_dates()
    lat, lon = get_lat_lon() # Note: Lua script missed defining this in reserve(), but used it.
    hotel_id = random.randint(1, 80)
    user_id_num, _, password = get_user_creds()
    cust_name = user_id_num # Logic from Lua: cust_name = user_id
    num_room = "1"

    # Note: The Lua script constructs a path with params, sending an empty body POST
    url = (f"{BASE_URL}/reservation?inDate={in_date}"
           f"&outDate={out_date}&lat={lat}&lon={lon}"
           f"&hotelId={hotel_id}&customerName={cust_name}"
           f"&username={user_id_num}&password={password}"
           f"&number={num_room}")

    return f'curl -s -X POST "{url}"'

def generate_requests(num_requests=1000):
    """Generate a list of curl commands based on weighted task distribution"""
    
    # Weights derived from the commented logic in the Lua script
    # search: 0.6, recommend: 0.39, user: 0.005, reserve: 0.005
    task_weights = {
        'search_hotel': 600,
        'recommend': 390,
        'user_login': 5,
        'reserve': 5
    }
    
    total_weight = sum(task_weights.values()) # Should be 1000
    
    # Calculate number of requests for each task type
    tasks = []
    for task_name, weight in task_weights.items():
        count = int((weight / total_weight) * num_requests)
        tasks.extend([task_name] * count)
    
    # Fill remaining due to rounding
    while len(tasks) < num_requests:
        tasks.append('search_hotel')
        
    random.shuffle(tasks)
    
    curl_commands = []
    for task in tasks:
        if task == 'search_hotel':
            curl_commands.append(generate_search_hotel_curl())
        elif task == 'recommend':
            curl_commands.append(generate_recommend_curl())
        elif task == 'user_login':
            curl_commands.append(generate_user_login_curl())
        elif task == 'reserve':
            curl_commands.append(generate_reserve_curl())
            
    return curl_commands

def main():
    parser = argparse.ArgumentParser(description='Generate curl requests for Hotel Reservation')
    parser.add_argument('--num-requests', type=int, default=1000,
                       help='Number of requests to generate (default: 1000)')
    parser.add_argument('--base-url', type=str, default='http://localhost:11000',
                       help='Base URL for the application (default: http://localhost:11000)')
    parser.add_argument('--execute', action='store_true',
                       help='Execute the curl commands')
    parser.add_argument('--output', type=str,
                       help='Output file to save curl commands')
    
    args = parser.parse_args()
    
    global BASE_URL
    BASE_URL = args.base_url
    
    print(f"Generating {args.num_requests} curl requests for {BASE_URL}...")
    curl_commands = generate_requests(args.num_requests)
    
    print(f"Generated {len(curl_commands)} curl commands")
    
    if args.output:
        with open(args.output, 'w') as f:
            f.write("#!/bin/bash\n")
            for cmd in curl_commands:
                f.write(cmd + '\n')
        print(f"Curl commands saved to {args.output}")
        # Make executable
        subprocess.run(f"chmod +x {args.output}", shell=True)
    
    if args.execute:
        print("Executing curl commands...")
        success_count = 0
        fail_count = 0
        connection_errors = 0
        http_errors = 0
        
        for i, cmd in enumerate(curl_commands, 1):
            try:
                result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
                
                # Check curl exit code
                if result.returncode != 0:
                    fail_count += 1
                    connection_errors += 1
                    if fail_count <= 5:  # Show first 5 errors
                        print(f"[{i}] CURL ERROR (exit {result.returncode}): {result.stderr.strip()}")
                # Check for empty response or HTTP errors in response
                elif not result.stdout or result.stdout.strip() == "":
                    fail_count += 1
                    connection_errors += 1
                    if fail_count <= 5:
                        print(f"[{i}] EMPTY RESPONSE - server may be unreachable")
                elif "error" in result.stdout.lower() or "failed" in result.stdout.lower():
                    fail_count += 1
                    http_errors += 1
                    if fail_count <= 5:
                        print(f"[{i}] HTTP ERROR: {result.stdout[:200]}")
                else:
                    success_count += 1
                    
            except Exception as e:
                fail_count += 1
                if fail_count <= 5:
                    print(f"[{i}] EXCEPTION: {e}", file=sys.stderr)
            
            if i % 100 == 0:
                print(f"Progress: {i}/{len(curl_commands)} | Success: {success_count} | Failed: {fail_count}")
        
        print(f"\n{'='*50}")
        print(f"EXECUTION COMPLETE")
        print(f"{'='*50}")
        print(f"Total requests:     {len(curl_commands)}")
        print(f"Successful:         {success_count}")
        print(f"Failed:             {fail_count}")
        if connection_errors > 0:
            print(f"  - Connection errors: {connection_errors}")
        if http_errors > 0:
            print(f"  - HTTP errors:       {http_errors}")
        print(f"Success rate:       {100*success_count/len(curl_commands):.1f}%")
    else:
        # Print examples
        print("\nFirst 5 curl commands:")
        for i, cmd in enumerate(curl_commands[:5], 1):
            print(f"{i}. {cmd}")
        
        print(f"\nUse --execute to run all commands")
        print(f"Use --output <file> to save to a script")

if __name__ == '__main__':
    main()