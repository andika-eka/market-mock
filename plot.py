import requests
import pandas as pd
import matplotlib.pyplot as plt
import pprint

SYMBOL = "BTC-USD, ETH-USD, XRP-USD"
DAYS = 365 * 10
# DAYS =  2
#  
INTERVAL = "240h"

url = f"http://localhost:8080/api/candles?symbol={SYMBOL}&days={DAYS}&interval={INTERVAL}"

try:
    response = requests.get(url)
    response.raise_for_status() 
    
    api_response = response.json()
    
    pprint.pprint(api_response)
    
    if not api_response.get('success', False):
        print(f"API Error: {api_response.get('message')}")
        exit()

    data_list = api_response['data']
    
    plt.figure(figsize=(12, 6))
    
    for data in data_list:
        print(f"Processing symbol: {data["Symbol"]}")
        
        candles = data["Candles"]

        
        # for candle in candles:
            
        df = pd.DataFrame(data["Candles"])
        
        df['time'] = pd.to_datetime(df['time'])
        df['close'] = df['close'].astype(float)
        
        plt.plot(df['time'], df['close'], label=data["Symbol"], linewidth=1)

    plt.title(f"Market Simulation: {SYMBOL} ({DAYS} Days)")
    plt.xlabel("Date")
    plt.ylabel("Price ($)")
    plt.legend()
    plt.grid(True, alpha=0.3)
    plt.savefig("market_simulation.png")

except Exception as e:
    print(e)