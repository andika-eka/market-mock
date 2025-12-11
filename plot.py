import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv("market_data.csv")

df['Time'] = pd.to_datetime(df['Time'])

plt.figure(figsize=(12, 6))

unique_symbols = df['Symbol'].unique()

for sym in unique_symbols:
    data = df[df['Symbol'] == sym]
    
    plt.plot(data['Time'], data['Close'], label=sym)

plt.title("Deterministic Market Simulation (10 Year History)")
plt.xlabel("Year")
plt.ylabel("Price ($)")
plt.legend()
plt.grid(True, alpha=0.3)

plt.savefig("market_simulation.png")