import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv("market_data.csv")

df['date'] = pd.to_datetime(df['date'])

plt.figure(figsize=(12, 6))

unique_symbols = df['symbol'].unique()

for sym in unique_symbols:
    data = df[df['symbol'] == sym]
    
    plt.plot(data['date'], data['close'], label=sym)

plt.title("Deterministic Market Simulation (10 Year History)")
plt.xlabel("Year")
plt.ylabel("Price ($)")
plt.legend()
plt.grid(True, alpha=0.3)

plt.savefig("market_simulation.png")