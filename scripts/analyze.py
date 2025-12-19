import glob, statistics

def parse(path):
    r = {}
    for l in open(path):
        if "Elapsed" in l:
            r["t"] = float(l.split("=")[1][:-2])
        if "NumGC" in l:
            r["gc"] = int(l.split("=")[1])
        if "GC Pause total" in l:
            r["pause"] = float(l.split("=")[1].replace(" ms",""))
    return r

def avg(pattern):
    rs = [parse(p) for p in glob.glob(pattern)]
    return {
        "time": statistics.mean(r["t"] for r in rs),
        "pause": statistics.mean(r["pause"] for r in rs),
        "gc": statistics.mean(r["gc"] for r in rs),
    }

old = avg("logs/oldgc_*.log")
new = avg("logs/greentea_*.log")

print("OLD GC:", old)
print("GREEN-TEA GC:", new)

print("\nPAUSE REDUCTION:",
      (old["pause"] - new["pause"]) / old["pause"] * 100, "%")
