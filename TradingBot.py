# IMPORT MODULES
import pandas as pd
import numpy as np
import warnings 
warnings.filterwarnings("ignore")
import statsmodels.api as sm
from statsmodels.tsa.stattools import coint
from statsmodels.tsa.stattools import adfuller
import math
from typing import Any, Dict, List, Optional, Tuple, Union

TIMESTAMPS = 15000
N_ASSETS = 7

current_positions = [0]*7 # Read Only For Algorithm - Only Utilized By Backtester 
position_limits = [50]*7  # Maximum Long/Short Position Allowed Independently - Read Only Value
position_deltas =   [0]*7 # Updated By Algorithm Every Tick - Read Only For Backtester
PnL =   [0]*7 # Updated By Algorithm Every Tick - Read Only For Backtester

path_to_data = 'BOLT_Dataset_15000'

Orderbook_snapshots = []
Trade_snapshots = []

for asset in range(N_ASSETS):
    #Orderbook_snapshots.append(pd.read_csv(rf'{path_to_data}\OB_snaps_15000\Asset_{asset+1}_OB_snapshot_15000.csv'))
    #Trade_snapshots.append(pd.read_csv(rf'{path_to_data}\TD_snaps_15000\Asset_{asset+1}_TD_snapshot_15000.csv'))
    Orderbook_snapshots.append(pd.read_csv(f"OB_snaps_15000\Asset_{asset+1}_OB_snapshot_15000.csv"))
    Trade_snapshots.append(pd.read_csv(f"TD_snaps_15000\Asset_{asset+1}_TD_snapshot_15000.csv"))
#print(Orderbook_snapshots)
#print(Trade_snapshots)
print(len(Orderbook_snapshots))
# -----------------------------------------------------------------------------------------------
#--------------------------------Pairs Trading Tests on Old data---------------------------------
#------------------------------------------------------------------------------------------------


def create_featured_dataset(ob_df, td_df):
    # ---- Preprocess ----
    ob_df.rename(columns={'Unnamed: 0': 'timestamp'}, inplace=True)
    td_df.rename(columns={'Unnamed: 0': 'trade_id', 't': 'timestamp'}, inplace=True)

    # --- Target ---
    td_df['price_volume'] = td_df['price']*td_df['qty']
    vwap_by_time = td_df.groupby('timestamp').apply(lambda x: (x['price_volume'].sum()/x['qty'].sum()) if x['qty'].sum()>0 else   np.nan).reset_index(name='vwap')

    #ob_df = ob_df.merge(vwap_by_time, on='timestamp', how='left')
    ob_df["vwap"]=vwap_by_time["vwap"]

    ob_df["daily_return"]=ob_df["vwap"].pct_change()
    ob_df["daily_return"]=ob_df["daily_return"]*100
    ob_df['close']=ob_df['vwap']
    ob_df['mid_price'] = ob_df['close']
    ob_df['high']=ob_df['close'].rolling(window=7).max()
    ob_df['low']=ob_df['close'].rolling(window=7).min()
    
    size_cols = ['bid_sz1', 'ask_sz1', 'bid_sz2', 'ask_sz2', 'bid_sz3', 'ask_sz3']
    ob_df['total_volume_top3'] = ob_df[size_cols].sum(axis=1)
    volatility_50 = ob_df['close'].pct_change().rolling(window=50).std()
    ob_df['market_temp'] = volatility_50 * np.log1p(ob_df['total_volume_top3'])
    proportions = ob_df[size_cols].div(ob_df['total_volume_top3'], axis=0)
    entropy = -np.sum(proportions * np.log2(proportions.replace(0, 1e-10)), axis=1)
    ob_df['market_entropy'] = entropy   
    print()
    return ob_df


def pairs_trading_tests(N_Assets):
    Old_Orderbook = []
    Old_Tradebook = []

#---------------------------below is the loop for loading old datasets---------------------------
    for asset in range(N_ASSETS):
        #Old_Orderbook.append(pd.read_csv(rf'{path_to_data}\OB_snaps_15000\Asset_{asset+1}_OB_snapshot_15000.csv'))
        #Old_Tradebook.append(pd.read_csv(rf'{path_to_data}\TD_snaps_15000\Asset_{asset+1}_TD_snapshot_15000.csv'))
        Old_Orderbook.append(pd.read_csv(f"OB_snaps_15000\Asset_{asset+1}_OB_snapshot_15000.csv"))
        Old_Tradebook.append(pd.read_csv(f"TD_snaps_15000\Asset_{asset+1}_TD_snapshot_15000.csv"))

    for i in range(N_Assets):
        Old_Orderbook[i]=create_featured_dataset(Old_Orderbook[i],Old_Tradebook[i])
    

    returns_df=pd.DataFrame({"Asset_1":Old_Orderbook[0]["daily_return"],"Asset_2":Old_Orderbook[1]["daily_return"],"Asset_3":Old_Orderbook[2]["daily_return"],"Asset_4":Old_Orderbook[3]["daily_return"],"Asset_5":Old_Orderbook[4]["daily_return"],"Asset_6":Old_Orderbook[5]["daily_return"],"Asset_7":Old_Orderbook[6]["daily_return"]})
    corr_matrix=returns_df.corr()
    #print(corr_matrix)

    asset_names=[
    "Asset_1","Asset_2","Asset_3","Asset_4","Asset_5","Asset_6","Asset_7"
    ]
    asset_vwap={
        "Asset_1":Old_Orderbook[0]["vwap"],"Asset_2":Old_Orderbook[1]["vwap"],"Asset_3":Old_Orderbook[2]["vwap"],"Asset_4":Old_Orderbook[3]["vwap"],"Asset_5":Old_Orderbook[4]["vwap"],"Asset_6":Old_Orderbook[5]["vwap"],"Asset_7":Old_Orderbook[6]["vwap"]
    }
    key_list = list(asset_vwap.keys())
    
    p_val=[]
    scores=[]
    asset1=[] #asset1
    asset2=[] #asset2
    
    for i in range(7):
        for j in range(i+1,7):
            a1=asset_names[i]
            a2=asset_names[j]
            series1=asset_vwap[a1]
            series2=asset_vwap[a2]
            score,p_value,_=coint(series1,series2)
            
            p_val.append(p_value)
            scores.append(score)
            asset1.append(a1)
            asset2.append(a2)
    
    coint_df=pd.DataFrame({"p_value":p_val,"score":scores,"asset1":asset1,"asset2":asset2})

    coint_a1=[]# 1st asset in list of asset which have p-val<0.05 
    coint_a2=[]
    
    p_val_list=[]
    for i in range(len(coint_df)):
        if (coint_df.loc[i,"p_value"]<0.05):
            coint_a1.append(coint_df.loc[i,"asset1"])
            coint_a2.append(coint_df.loc[i,"asset2"])
            #print(coint_df.loc[i,"asset1"]+" and "+coint_df.loc[i,"asset2"])
    
    for i in range(len(coint_a1)):
        
        # Suppose we have two price series
        s1 = asset_vwap[coint_a1[i]]
        s2 = asset_vwap[coint_a2[i]]
        
        # Step 1: Regress s1 on s2
        s2_const = sm.add_constant(s2)
        model = sm.OLS(s1, s2_const).fit()
        hedge_ratio = model.params[1]  # beta
        alpha = model.params[0]
        
        # Step 2: Residuals (spread)
        residuals = s1 - (alpha + hedge_ratio * s2)
        
        # Step 3: ADF test on residuals
        adf_stat, p_value, _, _, crit_values, _ = adfuller(residuals)

        p_val_list.append(p_value)
    
        print(coint_a1[i]+" and "+coint_a2[i])
        print("Hedge ratio (beta):", hedge_ratio)
        print("ADF statistic:", adf_stat)
        print("p-value:", p_value)
        print("Critical values:", crit_values)
        
        if p_value < 0.05:
            print("Pairs Trading can be applied on "+coint_a1[0]+ " and "+ coint_a2[0])
            #print("Residuals are stationary so Cointegration confirmed")

    
        print()
        

    min_p_value_in_list=min(p_val_list)
    min_p_value_idx=p_val_list.index(min_p_value_in_list)
    print(min_p_value_idx)

    print()
    print("Pairs Trading will be applied on "+coint_a1[0]+" and "+coint_a2[0]+ " as they have min p_value")
    
    index_of_asset1 = key_list.index(coint_a1[0])
    index_of_asset2= key_list.index(coint_a2[0])
    return index_of_asset1,index_of_asset2

index_of_asset1,index_of_asset2=pairs_trading_tests(N_ASSETS)
    
#----------------------------------------------------------------------------------------------------
#---------------------------Strategy Applied on New Hidden Datasets----------------------------------
#----------------------------------------------------------------------------------------------------


#----------------------------------------------------------------------------------------------------
#--------------------------------------1. Pairs Trading ---------------------------------------------
#----------------------------------------------------------------------------------------------------



def pairs_trading_preprocessing(ob1,ob2,td1,td2):
    ob1=create_featured_dataset(ob1,td1)
    ob2=create_featured_dataset(ob2,td2)

    x=ob1["vwap"]
    y=ob2["vwap"]
    ratio=x/y
    lookback = 20  # rolling window for calculating mean and standard deviation
    pt_df=pd.DataFrame({"ratio":ratio})
    
    
    pt_df["ratio_mean"] = ratio.rolling(window=lookback).mean()
    
    pt_df["ratio_std"] = ratio.rolling(window=lookback).std()
    
    pt_df["z_score"] = (pt_df["ratio"] - pt_df["ratio_mean"]) / pt_df["ratio_std"]
    return pt_df

x=500
def generate_pairs_trading_signals(pt_df):
    signal1=[]
    signal2=[]
    lookback=20
    for i in range(lookback):
        signal1.append(0)
        signal2.append(0)

    entry_threshold=1
    exit_threshold=0
    pos=0
    for i in range(lookback,len(pt_df)):
        curr_z_score=pt_df.loc[i,"z_score"]

        if curr_z_score>entry_threshold and pos==0:
            signal1.append(-1)
            signal2.append(1)
            pos=1
        elif curr_z_score<-(entry_threshold) and pos==0:
            signal1.append(1)
            signal2.append(-1)
            pos=-1
        elif curr_z_score<exit_threshold and pos==1:
            signal1.append(1)
            signal2.append(-1)
            pos=0
        elif curr_z_score>(exit_threshold) and pos==-1:
            signal1.append(-1)
            signal2.append(1)
            pos=0
        else:
            signal1.append(0)
            signal2.append(0)
    #squaring off any last signals
    if pos==1:
        signal1[-1]=1
        signal2[-1]=-1
    elif pos==-1:
        signal1[-1]=-1
        signal2[-1]=1

    df1=pd.DataFrame({'t':pt_df.index,"signal":signal1})
    df2=pd.DataFrame({'t':pt_df.index,"signal":signal2})
    return df1,df2


class Backtester:
    def __init__(self,
                 orderbook: pd.DataFrame,
                 trades: pd.DataFrame,
                 pos_limit: int = 50,
                 brokerage_bps: float = 1.0,
                 tape_cap_mode: str = "total"  # "total" or "side"
                 ):
        self.pos_limit = int(pos_limit)
        self.brokerage_rate = float(brokerage_bps) * 1e-4  # bps -> decimal
        self.tape_cap_mode = tape_cap_mode  # "total" (sum) or "side" (aggressor-aware if side provided)

        ob = orderbook.copy()
        if "t" not in ob.columns:
            ob["t"] = ob.index
        

        # ---------- expand L2 → L3-like ladders per tick ----------
        recs = []
        for lvl in (1, 2, 3):
            apx, asz = f"ask_px{lvl}", f"ask_sz{lvl}"
            bpx, bsz = f"bid_px{lvl}", f"bid_sz{lvl}"
            if apx in ob.columns and asz in ob.columns:
                part = ob[["t", apx, asz]].rename(columns={apx: "price", asz: "qty"})
                part["side"] = "ask"
                recs.append(part[["t", "side", "price", "qty"]])
            if bpx in ob.columns and bsz in ob.columns:
                part = ob[["t", bpx, bsz]].rename(columns={bpx: "price", bsz: "qty"})
                part["side"] = "bid"
                recs.append(part[["t", "side", "price", "qty"]])
        if not recs:
            raise ValueError("Order book must contain top-3 level columns for bids and asks.")

        l3 = pd.concat(recs, ignore_index=True)
        l3["side"]  = l3["side"].astype(str).str.lower().str.strip()
        l3["price"] = pd.to_numeric(l3["price"], errors="coerce").astype(float)
        l3["qty"]   = pd.to_numeric(l3["qty"],   errors="coerce").astype(float)
        l3 = l3.dropna(subset=["price", "qty"])
        l3 = l3[l3["qty"] > 0]

        # aggregate duplicates at same price per tick
        grp = l3.groupby(["t", "side", "price"], as_index=False)["qty"].sum()
        asks = grp[grp["side"].eq("ask")].sort_values(["t", "price"]).groupby("t")
        bids = grp[grp["side"].eq("bid")].sort_values(["t", "price"], ascending=[True, False]).groupby("t")

        self._asks_by_t: Dict[Any, List[Tuple[float, float]]] = {
            t: [(float(p), float(q)) for p, q in g[["price", "qty"]].to_numpy()]
            for t, g in asks
        }
        self._bids_by_t: Dict[Any, List[Tuple[float, float]]] = {
            t: [(float(p), float(q)) for p, q in g[["price", "qty"]].to_numpy()]
            for t, g in bids
        }

        td = trades.copy()
        print(td.head())
        
        
        
        td["qty"]  = pd.to_numeric(td.get("qty", 0), errors="coerce").fillna(0.0).astype(float)
        td["side"] = td.get("side", "both")
        td["side"] = td["side"].astype(str).str.lower().str.strip()

        self._trades_by_t = dict(tuple(td.groupby("timestamp"))) if not td.empty else {}
        # global tick axis = union of ticks present in OB or trades
        ts_ob = pd.Index(self._asks_by_t.keys()).union(self._bids_by_t.keys())
        ts_td = pd.Index(self._trades_by_t.keys())
        self._ts =ob["t"].tolist()

    # -------------- Public API --------------
    def run_with_signals(self,
                         signals: Union[pd.DataFrame, pd.Series, Dict[Any, float], List[float], np.ndarray],
                         t_col: str = "t",
                         sig_col: str = "signal",
                         threshold: float = 0.0,
                         huge_order: int = 10**9,
                         return_traces: bool = False):
        
        sig_by_t = self._coerce_signals(signals, t_col, sig_col)
        pos, cash = 0.0, 0.0
        rows = []
        

        for t in (self._ts):
            sig = float(sig_by_t.get(t, 0.0))
            desired_delta = +huge_order if sig > threshold else (-huge_order if sig < -threshold else 0)

            fills, filled_signed = self._execute_delta(
                tick=t,
                desired_delta=desired_delta,
                position_before=pos
            )

            signed_notional = sum(s * q * px for (q, px,fees, s) in fills)
            fees_total      = sum(s * q * fees for (q, px,fees, s) in fills)
            cash -= signed_notional + fees_total
            pos  += filled_signed
            if return_traces:
              rows.append({
            "t": t,
            "signal": sig,
            "desired_delta": desired_delta,
            "filled_qty": filled_signed,
            "cash_after": cash,
            "position_after": pos,
            "total_pnl": cash
        })
        if pos!=0:
            flatten_delta=-pos
            pos+=flatten_delta
            return (
                 {"final_position": float(pos), "final_cash": float(cash), "final_total_pnl": float(cash)},
                 (pd.DataFrame(rows) if return_traces else None),
             )
            
        
    # f any remainder still exists (ran out of book depth), do 
          

        if return_traces:
            rows.append({
                "t": t_last,
                "signal": 0.0,
                        # forced fully
                "cash_after": cash,
                "position_after": pos,
                "total_pnl": cash
            })

# (optional) sanity check
# assert pos == 0, f"Not flat after square-off, pos={pos}"
        

    
    def _levels_for_tick(self, t: Any) -> Dict[str, List[Tuple[float, float]]]:
        return {
            "asks": self._asks_by_t.get(t, []),
            "bids": self._bids_by_t.get(t, []),
        }

    def _tape_cap(self, t: Any) -> float:

       trades_t = self._trades_by_t.get(t, None)
       if trades_t is None or trades_t.empty:
          return 0.0
       return float(pd.to_numeric(trades_t["qty"], errors="coerce").fillna(0.0).sum())


    def _consume_levels(self, levels: List[Tuple[float, float]], qty_signed: float):
        if qty_signed == 0 or not levels:
            return [], 0.0
        side_sign = +1 if qty_signed > 0 else -1
        remain = abs(float(qty_signed))
        filled = 0.0
        fills: List[Tuple[float, float,float,int]] = []
        for px, sz in levels:
            if remain <= 0:
                break
            hit = min(remain, float(sz))
            if hit <= 0:
                continue
            fees=abs(hit*self.brokerage_rate)
            fills.append((hit, float(px),fees, side_sign))
            filled += hit
            remain -= hit
        return fills, float(math.copysign(filled, qty_signed))

    def _execute_delta(self, tick, desired_delta, position_before, *, ignore_tape: bool =False):
    # position-limit clip (unchanged)
       max_buy  = self.pos_limit - position_before
       max_sell = - (self.pos_limit + position_before)
       if desired_delta > 0:   desired_delta = min(desired_delta, max_buy)
       elif desired_delta < 0: desired_delta = max(desired_delta, max_sell)
       if desired_delta == 0:  return [], 0.0

       levels = self._levels_for_tick(tick)
       ladder = levels["asks"] if desired_delta > 0 else levels["bids"]
       if not ladder: return [], 0.0

       if ignore_tape:
           clipped = desired_delta
       else:
           cap = self._tape_cap(tick)
           clipped = math.copysign(min(abs(desired_delta), cap), desired_delta)
           if clipped == 0: return [], 0.0

       fills, filled_signed = self._consume_levels(ladder, clipped)
       return fills, filled_signed


    def _coerce_signals(self, signals, t_col="t", sig_col="signal") -> Dict[Any, float]:
        ts = self._ts
        if isinstance(signals, pd.DataFrame):
            if t_col not in signals.columns or sig_col not in signals.columns:
                raise ValueError(f"signals DataFrame must have columns '{t_col}' and '{sig_col}'")
            s = signals[[t_col, sig_col]].copy()
            s[sig_col] = pd.to_numeric(s[sig_col], errors="coerce").fillna(0.0)
            return dict(zip(s[t_col].values, s[sig_col].values))
        if isinstance(signals, pd.Series):
            s = pd.to_numeric(signals, errors="coerce").fillna(0.0)
            if np.array_equal(s.index.values, np.array(ts)):
                return {t: float(v) for t, v in s.items()}
            if len(s) == len(ts):
                return {t: float(v) for t, v in zip(ts, s.values)}
            raise ValueError("signals Series index must equal tick axis or length must match.")
        if isinstance(signals, dict):
            return {k: float(v) for k, v in signals.items()}
        if isinstance(signals, (list, np.ndarray)):
            if len(signals) != len(ts):
                raise ValueError("Length of signals must match tick axis.")
            return {t: float(v) for t, v in zip(ts, signals)}
        raise TypeError("Unsupported signals type.")





pairs_trading_df=pairs_trading_preprocessing(Orderbook_snapshots[index_of_asset1],Orderbook_snapshots[index_of_asset2],Trade_snapshots[index_of_asset1],Trade_snapshots[index_of_asset2])
signal1,signal2=generate_pairs_trading_signals(pairs_trading_df)

bt5=Backtester(Orderbook_snapshots[index_of_asset1],Trade_snapshots[index_of_asset1])
summary5, traces = bt5.run_with_signals(signal1, threshold=0.0, huge_order=10**9, return_traces=True)
PnL[index_of_asset1]=summary5.get('final_total_pnl')


bt7=Backtester(Orderbook_snapshots[index_of_asset2],Trade_snapshots[index_of_asset2])
summary7, traces = bt7.run_with_signals(signal2, threshold=0.0, huge_order=10**9, return_traces=True)
PnL[index_of_asset2]=summary7.get("final_total_pnl")


#-----------------------------------------------------------------------------------------------------
#------------------------------2. Thermodynamics Strategy on remaining assets-------------------------
#-----------------------------------------------------------------------------------------------------

# Orderbook_snapshots.pop(index_of_asset2)
# Orderbook_snapshots.pop(index_of_asset2)
# Trade_snapshots.pop(index_of_asset2)
# Trade_snapshots.pop(index_of_asset2)
# print(len(Orderbook_snapshots))
assets_ob = Orderbook_snapshots
assets_td = Trade_snapshots
N_assets = len(assets_ob)
#----------------------------------------------------------------------------------------
#----------------------PG FOR THRESHOLD OPTIMIZATION------------------------------------
#----------------------------------------------------------------------------------------
import numpy as np
import pandas as pd

def _discrete_mi(x_labels: pd.Series, y_labels: pd.Series) -> float:
    """
    Mutual Information between two discrete series (no sklearn).
    x_labels, y_labels already discretized/categorical.
    """
    df = pd.crosstab(x_labels, y_labels, normalize=True)
    px = df.sum(axis=1).values  # P(x)
    py = df.sum(axis=0).values  # P(y)
    mi = 0.0
    for i, xi in enumerate(df.index):
        for j, yj in enumerate(df.columns):
            pxy = df.iloc[i, j]
            if pxy > 0:
                mi += pxy * np.log(pxy / (px[i] * py[j]))
    return float(mi)

def _future_return_sign(mid_price: pd.Series, horizon: int = 1, deadband: float = 0.0) -> pd.Series:
    """
    Next-step return sign {-1,0,1}. deadband sets |ret|<deadband to 0.
    """
    r = mid_price.pct_change().shift(-horizon)  # future return
    y = r.copy()
    y[:] = 0
    y[r >  deadband] = 1
    y[r < -deadband] = -1
    return y

def _label_regime(temp: pd.Series, entropy: pd.Series, t_thr: float, e_thr: float) -> pd.Series:
    # Same quadrants as your get_market_regime()
    lab = pd.Series(index=temp.index, dtype=object)
    hot = temp > t_thr
    ord_ = entropy < e_thr
    lab[ hot &  ord_] = "Hot & Ordered"
    lab[~hot &  ord_] = "Cold & Ordered"
    lab[ hot & ~ord_] = "Hot & Disordered"
    lab[~hot & ~ord_] = "Cold & Disordered"
    return lab

def _objective_mi(df: pd.DataFrame, t_thr: float, e_thr: float,
                  horizon: int = 1, deadband: float = 0.0) -> float:
    """
    Objective to maximize: MI(regime, next-return-sign) - small penalties.
    """
    # Build labels
    regime = _label_regime(df['market_temp'], df['market_entropy'], t_thr, e_thr)
    y = _future_return_sign(df['mid_price'], horizon=horizon, deadband=deadband)

    # Align and drop NaNs
    idx = regime.index.intersection(y.index)
    regime = regime.loc[idx]
    y = y.loc[idx]

    # Guard small samples
    if len(idx) < 200:
        return -1e9

    # MI term
    mi = _discrete_mi(regime, y)

    # Class balance penalty: avoid degenerate splits (all in one bucket)
    counts = regime.value_counts(normalize=True)
    balance = 1.0 - np.sum(counts**2)  # Gini complement (higher is better)
    # Entropy of class distribution (optional extra smoothness)
    eps = 1e-12
    ent = -np.sum(counts * np.log(counts + eps))

    score = mi + 0.10 * balance + 0.05 * ent
    return float(score)

def _pg_optimize_for_asset(df: pd.DataFrame,
                            n_particles: int = 24,
                            iters: int = 40,
                            w: float = 0.70, c1: float = 1.5, c2: float = 1.5,
                            quantile_bounds=(0.05, 0.95),
                            horizon: int = 1,
                            deadband: float = 0.0,
                            rng_seed: int = 42,
                            objective="mi",
                            backtest_callback=None) -> tuple[float, float, float]:
    """
    PSO in 2D (t_thr, e_thr).
    objective:
        - "mi": maximize mutual information between regime and next-return sign.
        - "backtest": use backtest_callback(df, t_thr, e_thr) -> score (higher better).
    Returns: (best_score, best_t_thr, best_e_thr)
    """
    rng = np.random.default_rng(rng_seed)

    # Search box from empirical quantiles (robust)
    t_lo, t_hi = df['market_temp'].quantile(quantile_bounds[0]), df['market_temp'].quantile(quantile_bounds[1])
    e_lo, e_hi = df['market_entropy'].quantile(quantile_bounds[0]), df['market_entropy'].quantile(quantile_bounds[1])

    # Initialize particles uniformly in box
    X = np.empty((n_particles, 2))
    X[:, 0] = rng.uniform(t_lo, t_hi, size=n_particles)  # temp thr
    X[:, 1] = rng.uniform(e_lo, e_hi, size=n_particles)  # entropy thr
    V = rng.normal(scale=0.1, size=(n_particles, 2))

    def eval_particle(pos):
        t_thr, e_thr = float(pos[0]), float(pos[1])
        if objective == "backtest":
            if backtest_callback is None:
                return -1e9
            return float(backtest_callback(df, t_thr, e_thr))
        # default MI objective
        return _objective_mi(df, t_thr, e_thr, horizon=horizon, deadband=deadband)

    # Personal / global bests
    pbest = X.copy()
    pbest_val = np.array([eval_particle(x) for x in X])
    g_idx = int(np.argmax(pbest_val))
    gbest = pbest[g_idx].copy()
    gbest_val = float(pbest_val[g_idx])

    for _ in range(iters):
        r1 = rng.random((n_particles, 2))
        r2 = rng.random((n_particles, 2))
        # update velocity & position
        V = w*V + c1*r1*(pbest - X) + c2*r2*(gbest - X)
        X = X + V
        # clamp to box
        X[:, 0] = np.clip(X[:, 0], t_lo, t_hi)
        X[:, 1] = np.clip(X[:, 1], e_lo, e_hi)
        # evaluate
        vals = np.array([eval_particle(x) for x in X])
        better = vals > pbest_val
        pbest[better] = X[better]
        pbest_val[better] = vals[better]
        g_idx = int(np.argmax(pbest_val))
        if pbest_val[g_idx] > gbest_val:
            gbest_val = float(pbest_val[g_idx]); gbest = pbest[g_idx].copy()

    return gbest_val, float(gbest[0]), float(gbest[1])

def pg_thresholds_for_all_assets(featured_datasets: dict,
                                  n_particles: int = 24,
                                  iters: int = 10,
                                  horizon: int = 1,
                                  deadband: float = 0.0,
                                  objective: str = "mi",
                                  backtest_callback=None,
                                  verbose: bool = True) -> dict:
    """
    Run PSO per asset to get (temp_thr, entropy_thr).
    If objective == 'backtest', supply backtest_callback(df, t_thr, e_thr)->score.
    Returns: {asset_name: {'temp_thr':..., 'entropy_thr':..., 'score':...}}
    """
    results = {}
    for asset_name, df in featured_datasets.items():
        if verbose:
            print(f"\n[PSO] Optimizing thresholds for {asset_name}…")

        score, t_thr, e_thr = _pg_optimize_for_asset(
            df=df,
            n_particles=n_particles, iters=iters,
            horizon=horizon, deadband=deadband,
            objective=objective,
            backtest_callback=backtest_callback
        )
        results[asset_name] = {'temp_thr': t_thr, 'entropy_thr': e_thr, 'score': score}
        if verbose:
            print(f"  best_score={score:.5f}  temp_thr={t_thr:.6f}  entropy_thr={e_thr:.6f}")
    return results

#-------------------------------------------------------------
# ---------- Hurst exponent (of Variogram Approach) ----------
#-------------------------------------------------------------
def _hurst_exp(x: pd.Series, lags=(2, 4, 8, 16, 32, 64)) -> float:
    x = pd.Series(x).dropna().values
    if len(x) < max(lags) + 2: return np.nan
    taus, L = [], []
    for lag in lags:
        diff = x[lag:] - x[:-lag]
        sd = np.std(diff)
        if sd <= 0 or np.isnan(sd): continue
        taus.append(np.log(sd)); L.append(np.log(lag))
    if len(taus) < 2: return np.nan
    H = np.polyfit(L, taus, 1)[0]
    return float(np.clip(H, 0.0, 1.0))

def rolling_hurst(prices: pd.Series, window=400, lags=(2,4,8,16,32,64)) -> pd.Series:
    logp = np.log(pd.Series(prices).astype(float).replace(0, np.nan))
    return logp.rolling(window).apply(lambda w: _hurst_exp(w, lags=lags), raw=False)

#-------------------------------------------------------------
# --------------------- REGIME-WISE STRATEGY -----------------------
#-------------------------------------------------------------

# ---------- SMA X VORTEX (trend following) ----------
def sma(series: pd.Series, window: int = 20) -> pd.Series:
    return series.rolling(window).mean()

def vortex(df: pd.DataFrame, n: int = 14) -> pd.Series:
    """
    Needs high, low, close columns.
    Vortex Indicator (VI+ - VI-).
    """
    high, low, close = df['high'], df['low'], df['close']
    tr = (high.combine(close.shift(1), max) - low.combine(close.shift(1), min)).fillna(0)
    vm_plus  = (high - low.shift(1)).abs()
    vm_minus = (low - high.shift(1)).abs()
    vi_plus  = vm_plus.rolling(n).sum() / tr.rolling(n).sum()
    vi_minus = vm_minus.rolling(n).sum() / tr.rolling(n).sum()
    return (vi_plus - vi_minus).fillna(0.0)

def trend_signal(df: pd.DataFrame, sma_window=20, vortex_n=14) -> pd.Series:
    sma_line = sma(df['close'], sma_window)
    vortex_line = vortex(df, vortex_n)
    # signal: sign of SMA × Vortex (amplifies when both agree)
    return np.sign(sma_line * vortex_line).fillna(0).astype(int)


# ---------- CCI X MOMENTUM OSCILLATOR (mean reversion) ----------
def cci(df: pd.DataFrame, n: int = 20) -> pd.Series:
  
    tp = (df['high'] + df['low'] + df['close']) / 3
    sma_tp = tp.rolling(n).mean()
    mad = (tp - sma_tp).abs().rolling(n).mean()
    return (tp - sma_tp) / (0.015 * mad + 1e-12)

def momentum(series: pd.Series, n: int = 10) -> pd.Series:
    return series / series.shift(n) - 1

def meanrev_signal(df: pd.DataFrame, cci_n=20, mom_n=10) -> pd.Series:
    cci_line = cci(df, cci_n)
    mom_line = momentum(df['close'], mom_n)
    sig = np.sign(cci_line * mom_line).fillna(0).astype(int)
    return sig

# ------------------NORMALIZATION & SIZING--------------------------------

def fit_thermo_scalers(train_features: dict) -> dict:
    """Per-asset quantiles computed on TRAIN only."""
    scalers = {}
    for name, df in train_features.items():
        t = pd.to_numeric(df['market_temp'], errors='coerce')
        e = pd.to_numeric(df['market_entropy'], errors='coerce')
        scalers[name] = {
            't05': float(t.quantile(0.05)),
            't95': float(t.quantile(0.95)),
            'e05': float(e.quantile(0.05)),
            'e95': float(e.quantile(0.95)),
        }
    return scalers

def _norm01_with_scaler(series: pd.Series, lo: float, hi: float) -> pd.Series:
    return ((pd.to_numeric(series, errors='coerce') - lo) / ((hi - lo) + 1e-12)).clip(0, 1)

def thermo_risk_from_scaler(df: pd.DataFrame, sc: dict) -> pd.Series:
    t_n = _norm01_with_scaler(df['market_temp'], sc['t05'], sc['t95'])
    e_n = _norm01_with_scaler(df['market_entropy'], sc['e05'], sc['e95'])
    return (0.25 + 0.75 * (t_n * (1.0 - e_n))).fillna(0.25)

def assign_regime_from_thermo(df: pd.DataFrame,thr_dict : dict) -> pd.Series:
    t = pd.to_numeric(df['market_temp'], errors='coerce')
    e = pd.to_numeric(df['market_entropy'], errors='coerce')
    # t_hi, t_lo = t.quantile(0.70), t.quantile(0.30)
    # e_hi, e_lo = e.quantile(0.70), e.quantile(0.30)
    labels = []
    temp = thr_dict['temp_thr']
    entr = thr_dict['entropy_thr']
    for ti, ei in zip(t, e):
        if np.isnan(ti) or np.isnan(ei):
            labels.append("Neutral"); continue
        if ti >= temp and ei <= entr:
            labels.append("Hot & Ordered")
        elif ti <= temp and ei <= entr:
            labels.append("Cold & Ordered")
        elif ti >= temp and ei >= entr:
            labels.append("Hot & Disordered")
        elif ti <= temp and ei >= entr:
            labels.append("Cold & Disordered")
        else:
            labels.append("Neutral")              
    return pd.Series(labels, index=df.index)


# ------------------ STRATEGY IMPLEMENTATION ------------------
DESIRED_POSITION_SIZE = 10  # base size before thermo scaling
def apply_bot_strategy_hurst(featured_datasets: dict, thr_dict: dict,
                             hurst_window: int,
                             H_trend: float = 0.5,
                             H_mr: float = 0.4,
                             sma_window: int = 30,
                             vortex_n: int = 14,
                             cci_n: int = 20,
                             mom_n: int = 10) -> dict:
    """
    For each asset df (must contain 'timestamp' and 'mid_price' or 'price'):
      - rolling Hurst on price
      - market thermodynamics (temperature & entropy) for gating + risk sizing
      - Trend (H>=H_trend): sign( ΔSMA * Vortex )
      - Mean-rev (H<=H_mr): -sign( CCI * Momentum )
      - df['signal_pos'] ∈ {-1,0,1}, df['signal'] is DELTA units (already sized)
    """
    for name, df in featured_datasets.items():
        df = df.sort_values('timestamp').reset_index(drop=True)
        px = df['mid_price'] if 'mid_price' in df.columns else df['price']

        # Ensure OHLC columns exist for vortex/cci; fallbacks use close=px
        df['close'] = px
        df['high']=df['close'].rolling(window=7).max()
        df['low']=df['close'].rolling(window=7).min()

        # --- features
        df['hurst'] = rolling_hurst(px, window=hurst_window)
        df['risk_mult'] = df['thermo_risk'] if 'thermo_risk' in df.columns else thermo_risk_from_scaler(df)

        # --- regime label (reuse if present; else derive from thermo)
        if 'market_regime' not in df.columns:
            df['market_regime'] = assign_regime_from_thermo(df,thr_dict)

        reg = df['market_regime'].astype(str)
        gate_trend = ((reg == "Hot & Ordered") | (reg == "Cold & Ordered")) & (df['hurst'] >= H_trend)
        gate_mr    = (reg == "Cold & Ordered") & (df['hurst'] <= H_mr)

        # OPTIONAL: structural gating by entropy imbalance if available
        if 'entropy_imbalance' in df.columns:
            ei = df['entropy_imbalance'].astype(float)
            gate_trend = gate_trend & (ei > 0.0)
            gate_mr    = gate_mr    & (ei < 0.0)

        # --- Trend-following: SMA × Vortex (use SMA slope)
        sma_line   = sma(df['close'], window=sma_window)
        sma_delta  = sma_line.diff()  # slope of SMA
        vortex_line = vortex(df, n=vortex_n)
        trend_score = (sma_delta * vortex_line)
        pos_tr = np.sign(trend_score).where(gate_trend, 0).fillna(0).astype(int)

        # --- Mean-reversion: CCI × Momentum (contrarian sign)
        cci_line = cci(df, n=cci_n)
        mom_line = momentum(df['close'], n=mom_n)
        mr_score = -(cci_line * mom_line)            # contrarian to joint overbought/over-sold with momentum
        pos_mr = np.sign(mr_score).where(gate_mr, 0).fillna(0).astype(int)

        # --- Combine (trend wins on ties)
        combined_pos = pos_tr.copy()
        both = (pos_tr != 0) & (pos_mr != 0)
        combined_pos[both] = pos_tr[both]
        combined_pos[combined_pos == 0] = pos_mr[combined_pos == 0]
        df['signal_pos'] = combined_pos.astype(int)


        # --- Turn target position into delta signal, with thermo sizing
        base_units = DESIRED_POSITION_SIZE if 'DESIRED_POSITION_SIZE' in globals() else 10
        tgt_units  = (base_units * df['risk_mult']).round().astype(int) * df['signal_pos']
        df['tgt_units'] = tgt_units
        df['signal'] = tgt_units.diff().fillna(tgt_units).astype(int)
        # print(df['signal'])

        featured_datasets[name] = df
        # print(featured_datasets)

    # Quick console report
    # print("\n=== Hurst-based assignment (SMA×Vortex trend; CCI×Momentum MR) ===")
    for name, df in featured_datasets.items():
        h = df['hurst'].dropna()
        if not h.empty:
            print(f"{name}: median H={h.median():.3f}  trend%={(df['signal_pos']==1).mean():.1%}  mr%={(df['signal_pos']==-1).mean():.1%}")
    return featured_datasets

import pandas as pd
import numpy as np
import math
from typing import Any, Dict, List, Optional, Tuple, Union

class Backtester:
    """
    L3-only taker backtester driven by (signal, volume) per tick.
    Required L2 OB columns: bid_px1..3, bid_sz1..3, ask_px1..3, ask_sz1..3 (+ optional 't')
    Trades tape columns: t, qty  (optional 'side' for side-aware capping)
    """

    def __init__(self,
                 orderbook_l2: pd.DataFrame,
                 trades: pd.DataFrame,
                 pos_limit: int = 50,
                 brokerage_bps: float = 1.0,
                 *,
                 tape_cap_mode: str = "total"  # "total" or "side"
                 ):
        self.pos_limit = int(pos_limit)
        self.brokerage_rate = float(brokerage_bps) * 1e-4  # bps -> decimal
        self.tape_cap_mode = tape_cap_mode  # "total" (sum) or "side" (aggressor-aware if side provided)

        # ---------- normalize L2 ----------
        ob = orderbook_l2.copy()
        if "t" not in ob.columns:
            ob["t"] = ob.index

        # ---------- expand L2 → L3-like ladders per tick ----------
        recs = []
        for lvl in (1, 2, 3):
            apx, asz = f"ask_px{lvl}", f"ask_sz{lvl}"
            bpx, bsz = f"bid_px{lvl}", f"bid_sz{lvl}"
            if apx in ob.columns and asz in ob.columns:
                part = ob[["t", apx, asz]].rename(columns={apx: "price", asz: "qty"})
                part["side"] = "ask"
                recs.append(part[["t", "side", "price", "qty"]])
            if bpx in ob.columns and bsz in ob.columns:
                part = ob[["t", bpx, bsz]].rename(columns={bpx: "price", bsz: "qty"})
                part["side"] = "bid"
                recs.append(part[["t", "side", "price", "qty"]])
        if not recs:
            raise ValueError("Order book must contain top-3 level columns for bids and asks.")

        l3 = pd.concat(recs, ignore_index=True)
        l3["side"]  = l3["side"].astype(str).str.lower().str.strip()
        l3["price"] = pd.to_numeric(l3["price"], errors="coerce").astype(float)
        l3["qty"]   = pd.to_numeric(l3["qty"],   errors="coerce").astype(float)
        l3 = l3.dropna(subset=["price", "qty"])
        l3 = l3[l3["qty"] > 0]

        # aggregate duplicates at same price per tick
        grp = l3.groupby(["t", "side", "price"], as_index=False)["qty"].sum()
        asks = grp[grp["side"].eq("ask")].sort_values(["t", "price"]).groupby("t")
        bids = grp[grp["side"].eq("bid")].sort_values(["t", "price"], ascending=[True, False]).groupby("t")

        self._asks_by_t: Dict[Any, List[Tuple[float, float]]] = {
            t: [(float(p), float(q)) for p, q in g[["price", "qty"]].to_numpy()]
            for t, g in asks
        }
        self._bids_by_t: Dict[Any, List[Tuple[float, float]]] = {
            t: [(float(p), float(q)) for p, q in g[["price", "qty"]].to_numpy()]
            for t, g in bids
        }

        # ---------- normalize trades (tape cap) ----------
        td = trades.copy()
        if "t" not in td.columns:
            td["t"] = td.index
        td["qty"]  = pd.to_numeric(td.get("qty", 0), errors="coerce").fillna(0.0).astype(float)
        td["side"] = td.get("side", "both")
        td["side"] = td["side"].astype(str).str.lower().str.strip()

        self._trades_by_t = dict(tuple(td.groupby("t"))) if not td.empty else {}
        # global tick axis = union of ticks present in OB or trades
        ts_ob = pd.Index(self._asks_by_t.keys()).union(self._bids_by_t.keys())
        ts_td = pd.Index(self._trades_by_t.keys())
        self._ts = sorted(ts_ob.union(ts_td))

    # =============== Public APIs ===============

    def run_with_orders(self,
                        signals: Union[pd.DataFrame, pd.Series, Dict[Any, float], List[float], np.ndarray],
                        volumes: Optional[Union[pd.DataFrame, pd.Series, Dict[Any, float], List[float], np.ndarray]] = None,
                        *,
                        t_col: str = "t",
                        sig_col: str = "signal",
                        vol_col: str = "volume",
                        threshold: float = 0.0,
                        return_traces: bool = False,
                        enforce_flat_at_end: bool = True,
                        virtual_depth_on_flatten: bool = True):
        """
        Execute a signed schedule derived from (signal, volume).
        - desired_signed_qty(t) = sign(signal[t]) * abs(volume[t]) if |signal| > threshold else 0
        - Then clip by:
            (1) position limit (±pos_limit)
            (2) trade tape cap at that tick (total or side-aware)
        - Then execute across price ladder for that side.

        You may also pass a DataFrame with columns [t, signal, volume], or a Series/dict for each.
        """
        schedule = self._build_signed_schedule(signals, volumes, t_col, sig_col, vol_col, threshold)
        pos, cash = 0.0, 0.0
        rows = []

        for t in self._ts:
            wish = float(schedule.get(t, 0.0))  # signed desired trade for this tick
            if wish == 0.0:
                if return_traces:
                    rows.append({
                        "t": t, "wish": 0.0, "after_pos": pos, "after_cash": cash,
                        "filled_qty": 0.0, "vwap": np.nan, "fees": 0.0, "pnl_so_far": cash
                    })
                continue

            # --- constraint (1): position limit
            max_buy  = self.pos_limit - pos
            max_sell = - (self.pos_limit + pos)
            wish = min(wish, max_buy) if wish > 0 else max(wish, max_sell)
            if wish == 0.0:
                if return_traces:
                    rows.append({
                        "t": t, "wish": 0.0, "after_pos": pos, "after_cash": cash,
                        "filled_qty": 0.0, "vwap": np.nan, "fees": 0.0, "pnl_so_far": cash,
                        "note": "blocked_by_pos_limit"
                    })
                continue

            # --- constraint (2): tape cap
            cap = self._tape_cap(t, wish)
            wish = math.copysign(min(abs(wish), cap), wish)
            if wish == 0.0:
                if return_traces:
                    rows.append({
                        "t": t, "wish": 0.0, "after_pos": pos, "after_cash": cash,
                        "filled_qty": 0.0, "vwap": np.nan, "fees": 0.0, "pnl_so_far": cash,
                        "note": "blocked_by_tape_cap"
                    })
                continue

            # --- execute across ladder
            levels = self._levels_for_tick(t)
            ladder = levels["asks"] if wish > 0 else levels["bids"]
            fills, filled_signed = self._consume_levels(ladder, wish)

            # cash/pos update
            signed_notional = sum(s * q * px for (q, px, fee, s) in fills)
            fees_total      = sum(fee        for (q, px, fee, s) in fills)
            cash -= (signed_notional + fees_total)
            pos  += filled_signed

            vwap = (sum(q * px for (q, px, fee, s) in fills) / sum(q for (q, px, fee, s) in fills)) if fills else np.nan

            if return_traces:
                rows.append({
                    "t": t,
                    "wish": float(math.copysign(abs(wish), wish)),
                    "filled_qty": float(filled_signed),
                    "vwap": float(vwap) if not np.isnan(vwap) else np.nan,
                    "fees": float(fees_total),
                    "after_pos": float(pos),
                    "after_cash": float(cash),
                    "pnl_so_far": float(cash)
                })

        # --- flatten at last tick (ignore tape cap; optionally add virtual depth so you ALWAYS flatten)
        if enforce_flat_at_end and pos != 0:
            t_last = self._ts[-1]
            flatten_delta = -pos
            levels = self._levels_for_tick(t_last)
            ladder = levels["asks"] if flatten_delta > 0 else levels["bids"]

            if virtual_depth_on_flatten:
                fills, filled_signed = self._consume_levels_with_virtual(ladder, flatten_delta)
            else:
                fills, filled_signed = self._consume_levels(ladder, flatten_delta)

            signed_notional = sum(s * q * px for (q, px, fee, s) in fills)
            fees_total      = sum(fee        for (q, px, fee, s) in fills)
            cash -= (signed_notional + fees_total)
            pos  += filled_signed

            vwap = (sum(q * px for (q, px, fee, s) in fills) / sum(q for (q, px, fee, s) in fills)) if fills else np.nan

            if return_traces:
                rows.append({
                    "t": t_last,
                    "wish": float(flatten_delta),
                    "filled_qty": float(filled_signed),
                    "vwap": float(vwap) if not np.isnan(vwap) else np.nan,
                    "fees": float(fees_total),
                    "after_pos": float(pos),
                    "after_cash": float(cash),
                    "pnl_so_far": float(cash),
                    "note": "flatten"
                })

        return (
            {"final_position": float(pos), "final_cash": float(cash), "final_total_pnl": float(cash)},
            (pd.DataFrame(rows) if return_traces else None),
        )

    # Kept for backward-compatibility: if you ever still pass just signals (no volume),
    # it will behave like your old "huge order then clip" flow.
    def run_with_signals(self,
                         signals: Union[pd.DataFrame, pd.Series, Dict[Any, float], List[float], np.ndarray],
                         t_col: str = "t",
                         sig_col: str = "signal_pos",
                         threshold: float = 0.0,
                         huge_order: int = 10**9,
                         return_traces: bool = False):
        # Build schedule with implicit huge size, then clip by tape/pos/ladder.
        schedule = self._build_signed_schedule(signals, None, t_col, sig_col, vol_col="__none__", threshold=threshold,
                                               default_volume=huge_order)
        # Reuse the executor by passing a pre-signed schedule as dict
        return self.run_with_orders(pd.Series(schedule), volumes=None,
                                    t_col=t_col, sig_col=sig_col, vol_col="__none__",
                                    threshold=threshold,
                                    return_traces=return_traces)

    # =============== Internals ===============

    def _levels_for_tick(self, t: Any) -> Dict[str, List[Tuple[float, float]]]:
        return {"asks": self._asks_by_t.get(t, []), "bids": self._bids_by_t.get(t, [])}

    def _tape_cap(self, t: Any, desired_delta: float) -> float:
        """Total or side-aware cap using your trades."""
        trades_t = self._trades_by_t.get(t, None)
        if trades_t is None or trades_t.empty:
            return 0.0
        if self.tape_cap_mode == "side" and "side" in trades_t.columns:
            want_buy = desired_delta > 0
            side_mask = trades_t["side"].astype(str).str.lower().str.strip()
            mask = side_mask.isin(["buy", "b", "bid", "1", "+1"]) if want_buy else side_mask.isin(["sell", "s", "ask", "-1"])
            return float(pd.to_numeric(trades_t.loc[mask, "qty"], errors="coerce").fillna(0.0).sum())
        # total cap
        return float(pd.to_numeric(trades_t["qty"], errors="coerce").fillna(0.0).sum())

    def _consume_levels(self, levels: List[Tuple[float, float]], qty_signed: float):
        if qty_signed == 0 or not levels:
            return [], 0.0
        side_sign = +1 if qty_signed > 0 else -1
        remain = abs(float(qty_signed))
        filled = 0.0
        fills: List[Tuple[float, float, float, int]] = []
        for px, sz in levels:
            if remain <= 0:
                break
            hit = min(remain, float(sz))
            if hit <= 0:
                continue
            fee = abs(hit * float(px)) * self.brokerage_rate
            fills.append((hit, float(px), fee, side_sign))
            filled += hit
            remain -= hit
        return fills, float(math.copysign(filled, qty_signed))

    def _consume_levels_with_virtual(self, levels: List[Tuple[float, float]], qty_signed: float):
        """Consume real depth, then (if still remaining) fill remainder at the worst visible price (virtual infinite)."""
        fills, filled_signed = self._consume_levels(levels, qty_signed)
        want_buy = qty_signed > 0
        filled_abs = abs(filled_signed)
        remain = abs(qty_signed) - filled_abs
        if remain > 1e-12:
            if levels:
                worst_price = levels[-1][0] if want_buy else levels[-1][0]
            else:
                # if no ladder on that side: best available opposite side, else 0.0 (fallback)
                worst_price = 0.0
            side_sign = +1 if want_buy else -1
            fee = abs(remain * worst_price) * self.brokerage_rate
            fills.append((remain, worst_price, fee, side_sign))
            filled_signed = math.copysign(filled_abs + remain, qty_signed)
        return fills, filled_signed

    def _build_signed_schedule(self,
                               signals: Union[pd.DataFrame, pd.Series, Dict[Any, float], List[float], np.ndarray],
                               volumes: Optional[Union[pd.DataFrame, pd.Series, Dict[Any, float], List[float], np.ndarray]],
                               t_col: str, sig_col: str, vol_col: str, threshold: float,
                               default_volume: Optional[float] = None) -> Dict[Any, float]:
        """
        Returns dict[tick -> signed_qty].
        Accepted forms:
        - DataFrame with [t, signal, volume]
        - Two separate Series/arrays/dicts for signals and volumes
        - Only signals (+ default_volume) -> 'huge order' behavior
        """
        ts = self._ts

        # Helper to coerce any container into {t: value} aligned to self._ts
        def as_map(x, name: str) -> Dict[Any, float]:
            if x is None:
                return {}
            if isinstance(x, pd.DataFrame):
                if t_col not in x.columns:
                    x = x.reset_index().rename(columns={"index": t_col})
                col = sig_col if name == "sig" else vol_col
                if col not in x.columns:
                    raise ValueError(f"{name} DataFrame must have column '{col}'")
                v = pd.to_numeric(x[col], errors="coerce").fillna(0.0).values
                t = x[t_col].values
                return {tt: float(vv) for tt, vv in zip(t, v)}
            if isinstance(x, pd.Series):
                s = pd.to_numeric(x, errors="coerce").fillna(0.0)
                if np.array_equal(s.index.values, np.array(ts)):
                    return {t: float(v) for t, v in s.items()}
                if len(s) == len(ts):
                    return {t: float(v) for t, v in zip(ts, s.values)}
                return {t: float(v) for t, v in zip(ts[:len(s)], s.values)}
            if isinstance(x, dict):
                return {k: float(v) for k, v in x.items()}
            if isinstance(x, (list, np.ndarray)):
                if len(x) != len(ts):
                    # allow prefix length; align from start
                    return {t: float(v) for t, v in zip(ts[:len(x)], x)}
                return {t: float(v) for t, v in zip(ts, x)}
            raise TypeError(f"Unsupported type for {name} input.")

        sig_map = as_map(signals, "sig")
        if volumes is not None:
            vol_map = as_map(volumes, "vol")
        else:
            if default_volume is None:
                raise ValueError("volumes is None and default_volume not provided.")
            vol_map = {t: float(default_volume) for t in ts}

        schedule: Dict[Any, float] = {}
        for t in ts:
            s = float(sig_map.get(t, 0.0))
            v = abs(float(vol_map.get(t, 0.0)))
            if abs(s) <= threshold or v == 0.0:
                schedule[t] = 0.0
            else:
                side = 1.0 if s > 0 else -1.0
                schedule[t] = side * v
        return schedule
    
print(assets_ob)
def run_rolling_pg_backtest(
    assets_ob,  # Dict[str, DataFrame] OR List[DataFrame]
    assets_td,  # Dict[str, DataFrame] OR List[DataFrame]
    *,
    TRAIN_LEN: int = 700,
    TEST_LEN: int  = 600,
    # PSO/search params
    pso_particles: int = 24,
    pso_iters: int = 10,
    pso_horizon: int = 1,
    pso_deadband: float = 0.0,
    # Strategy params (Hurst+Thermo block)
    hurst_window: int = 300,
    H_trend: float = 0.55,
    H_mr: float = 0.45,
    sma_window: int = 20,
    vortex_n: int = 14,
    cci_n: int = 20,
    mom_n: int = 10,
    # Backtester params
    pos_limit: int = 50,
    brokerage_bps: float = 1.0,
    tape_cap_mode: str = "total",
    # Execution/verbosity
    return_traces: bool = True,
    verbose: bool = True,
    # Exclusions (by name). Defaults to skipping Asset_5 & Asset_6
    exclude_assets=("Asset_5", "Asset_6"),
    # Optional: exclude by 1-based index if you pass lists (e.g., (5,6))
    exclude_indices=(),
):
    """
    Walk-forward per asset (except those excluded):
      1) Build full features
      2) Slide TRAIN→TEST windows (no leakage)
      3) PSO on TRAIN -> thresholds
      4) Label TEST, compute thermo risk (fit on TRAIN, apply on TEST)
      5) Strategy -> signal_pos, tgt_units on TEST
      6) Backtest with (signal, volume) executor & constraints

    Returns:
      {
        "per_asset": {
            name: {"summary": dict, "trace": DataFrame|None, "signals": DataFrame},
            ...
        },
        "features_full": {name: DataFrame, ...},
        "skipped_assets": [names...]
      }
    """
    lost = {'Asset_1':0,'Asset_2':1,'Asset_3':2,'Asset_4':3,'Asset_5':4,'Asset_6':5,'Asset_7':6}
    # -------- helpers --------
    def _to_asset_dict(x, base="Asset"):
        # Accept dict or list/tuple; convert lists to {Asset_1: df, ...}
        if isinstance(x, dict):
            return {str(k): v for k, v in x.items()}
        if isinstance(x, (list, tuple)):
            return {f"{base}_{i+1}": v for i, v in enumerate(x)}
        raise TypeError("assets_ob/assets_td must be dict or list/tuple of DataFrames")

    def _ensure_t_col(df: pd.DataFrame) -> pd.DataFrame:
        df = df.copy()
        if "t" not in df.columns:
            if "timestamp" in df.columns:
                df["t"] = pd.to_numeric(df["timestamp"], errors="coerce").astype("Int64")
            else:
                df["t"] = pd.to_numeric(df.index, errors="coerce").astype("Int64")
        else:
            df["t"] = pd.to_numeric(df["t"], errors="coerce").astype("Int64")
        return df

    # -------- normalize inputs (this fixes your .keys() error) --------
    ob_dict = _to_asset_dict(assets_ob, "Asset")
    td_dict = _to_asset_dict(assets_td, "Asset")
    print(ob_dict.keys())
    # Align by common names; warn if some are missing in either side
    ob_names = set(ob_dict.keys())
    td_names = set(td_dict.keys())
    common = ('Asset_1', 'Asset_2', 'Asset_3', 'Asset_4', 'Asset_5','Asset_6','Asset_7')

    missing_in_ob = sorted(td_names - ob_names)
    missing_in_td = sorted(ob_names - td_names)
    if verbose and (missing_in_ob or missing_in_td):
        print(f"Warning: skipping assets without both OB & TD. missing_in_ob={missing_in_ob}, missing_in_td={missing_in_td}")

    # Exclusions
    excl = set(map(str, exclude_assets or ()))
    if exclude_indices:
        # If you passed lists and want to skip by 1-based index (5 -> Asset_5)
        for idx in exclude_indices:
            excl.add(f"Asset_{int(idx)}")

    selected_names = [n for n in common if n not in excl]
    if verbose:
        print(f"\nSelected assets (excluding {sorted(excl)}): {selected_names}")

    # -------- (0) Build full features for selected only --------
    if verbose: print("\nEngineering features (FULL) for selected assets…")
    features_full = {}
    for name in selected_names:
        df = create_featured_dataset(ob_dict[name], td_dict[name])
        df = df.copy()
        df['timestamp'] = pd.to_numeric(df['timestamp'], errors='coerce')
        df = df.dropna(subset=['timestamp']).sort_values('timestamp').reset_index(drop=True)
        features_full[name] = df

    results = {
        "per_asset": {},
        "features_full": features_full,
        "skipped_assets": sorted(excl),
    }

    # -------- (1) Roll, signal, backtest per asset --------
    for asset_name, feat in features_full.items():
        if verbose:
            print(f"\n=== Rolling PSO walk-forward for {asset_name} (train={TRAIN_LEN}, test={TEST_LEN}) ===")

        feat = feat.copy()
        n = len(feat)
        if n < TRAIN_LEN + 1:
            if verbose:
                print(f"  Skipping {asset_name}: not enough rows ({n}).")
            continue

        # Master signal frame on Int64 timestamps (zeros default)
        ts_idx = pd.to_numeric(feat['timestamp'], errors='coerce').astype('Int64').dropna()
        sig_series = pd.DataFrame(index=ts_idx)
        sig_series['signal_pos'] = 0
        sig_series['tgt_units']  = 0

        t = TRAIN_LEN
        while t < n:
            tr_start = max(0, t - TRAIN_LEN)
            tr_end   = t
            te_end   = min(n, t + TEST_LEN)

            train_slice = feat.iloc[tr_start:tr_end].copy()
            test_slice  = feat.iloc[t:te_end].copy()
            if len(train_slice) < TRAIN_LEN or len(test_slice) == 0:
                break

            # (1) PSO on TRAIN (no leakage)
            thr_dict = pg_thresholds_for_all_assets(
                featured_datasets={asset_name: train_slice},
                n_particles=pso_particles, iters=pso_iters,
                horizon=pso_horizon, deadband=pso_deadband,
                verbose=False
            )
            thr = thr_dict[asset_name]  # {'temp_thr':..., 'entropy_thr':...}

            # (2) Label TEST slice
            test_slice = test_slice.copy()
            test_slice['market_regime'] = _label_regime(
                test_slice['market_temp'],
                test_slice['market_entropy'],
                thr['temp_thr'], thr['entropy_thr']
            )

            # (3) Thermo risk: fit on TRAIN ONLY, apply on TEST
            scalers = fit_thermo_scalers({asset_name: train_slice})
            test_slice['thermo_risk'] = thermo_risk_from_scaler(test_slice, scalers[asset_name])

            # (4) Strategy -> signals on TEST
            tmp = apply_bot_strategy_hurst(
                featured_datasets={asset_name: test_slice.copy()},thr_dict={asset_name: thr},
                hurst_window=hurst_window,
                H_trend=H_trend,
                H_mr=H_mr,
                sma_window=sma_window,
                vortex_n=vortex_n,
                cci_n=cci_n,
                mom_n=mom_n
            )
            te_out = tmp[asset_name][['timestamp', 'signal_pos', 'tgt_units']].dropna()
            te_out['timestamp'] = pd.to_numeric(te_out['timestamp'], errors='coerce')
            te_out = te_out.dropna(subset=['timestamp'])
            te_idx = te_out['timestamp'].astype('Int64')

            overlap = sig_series.index.intersection(te_idx)
            if len(overlap) > 0:
                wr = te_out.set_index(te_idx).loc[overlap]
                sig_series.loc[overlap, 'signal_pos'] = wr['signal_pos'].astype('int64')
                sig_series.loc[overlap, 'tgt_units']  = wr['tgt_units'].astype('int64')

            t += TEST_LEN

        # Backtest (signal, volume)
        ob = _ensure_t_col(ob_dict[asset_name])
        td = _ensure_t_col(td_dict[asset_name])
        bt = Backtester(ob, td, pos_limit=pos_limit, brokerage_bps=brokerage_bps, tape_cap_mode=tape_cap_mode)

        bt_ticks = pd.Index(bt._ts)
        try:
            bt_ticks_int = pd.to_numeric(bt_ticks, errors='raise').astype('Int64')
            sig_aligned = sig_series.reindex(bt_ticks_int).fillna(0)
            sig_pos = pd.Series(sig_aligned['signal_pos'].astype(float).values, index=bt_ticks)
            tgt_vol = pd.Series(sig_aligned['tgt_units'].astype(float).values, index=bt_ticks)
        except Exception:
            sig_aligned = sig_series.reindex(bt_ticks).fillna(0)
            sig_pos = sig_aligned['signal_pos'].astype(float)
            tgt_vol = sig_aligned['tgt_units'].astype(float)

        summary, trace = bt.run_with_orders(
            signals=sig_pos,
            volumes=tgt_vol,
            return_traces=return_traces,
            enforce_flat_at_end=True,
            virtual_depth_on_flatten=True
        )

        # print(PnL)
        PnL[lost[asset_name]] = summary['final_total_pnl']


    return PnL

Final_PnL = run_rolling_pg_backtest(
    assets_ob=assets_ob,
    assets_td=assets_td,
    TRAIN_LEN=700,
    TEST_LEN=600,
    pos_limit=50,
    brokerage_bps=1.0,
    tape_cap_mode="total",
    return_traces=True,
    verbose=True
)

for i in range(7):
    print(f"Asset_{i+1} PnL: {Final_PnL[i]}")

print("Total PnL:", sum(Final_PnL.values()))


'''
for TICK in range(TIMESTAMPS):
    
    # ALGORITHM HERE
    # Use Orderbook snapshots of [0, TICK] to update position_deltas
    # current_positions = [...] Read Only For Algorithm
    # position_deltas = [...] Updated by algorithm evert TICK
    
    # Algorithm must not use Trade_snapshots data here
    
    pass

    # BACKTESTER HERE
    # Use Trade_snapshots with position_deltas to update current_positions and compute PnL asset-wise
    # CONSTRAINTS:
    #   position change must not exceed position_deltas
    #   position change must not exceed maximum available liquidity in Trade_snapshots at each level
    #   trades must only roll off to L2 and beyond if and only if previous levels are fully consumed
    #   position change must not cause current_positions exceed position_limits
    
    #  current_positions = [...] Updated by Backtester
    
    # current_positions = [...] Updated by Backtester
    # asset-wise PnL = [...]  Updated based on realized price and volume from trade_snapshots
    
    pass


# Comments and overall aggregate PnL
    
'''    
