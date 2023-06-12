import matplotlib.pyplot as plt

# Measurements from throughput.data

# Measurements from latency.data

bsize = [
    ('phalanx reordered rate', [
        (0, 0),  # 0
        (1, 0),  # 0
        (2, 0),  # 0
        (3, 0),  # 0
        (4, 0),  # 0.028
        (5, 0),  # 0.035
        (6, 0),  # 0.04
        (7, 0.09),  # 0.03
        (8, 0.16),  # 0.31
        (9, 0.1),  # 0.28
        (10, 0.32),  # 0.46
        (11, 0.59),  # 0.78
    ], '-p', 'brown'),
    ('timestamp-based reordered rate', [
        (0, 0),  # 0
        (1, 0.4),  # 1.9
        (2, 4.2),  # 3.9
        (3, 3.4),  # 5.2
        (4, 4.8),  # 7.64
        (5, 5.7),  # 8.28
        (6, 5.57),  # 10.14
        (7, 12),  # 14
        (8, 20.6),  # 18.89
        (9, 16.87),  # 17.87
        (10, 16.7),  # 15.1
        (11, 21.9),  # 19.0
    ], '-d', 'peru'),
]


def do_plot():
    f, ax = plt.subplots(1, 1, figsize=(6, 4))
    replicaNo = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
    xticks = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
    xticks_label = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"]
    thru = [
        ('Phalanx strategy', [
            [0.0, 0.0],
            [0.0, 0.0],
            [0.0, 0.0],
            [0.0, 0.0],
            [0.0, 0.028],
            [0.0, 0.035],
            [0.0, 0.040],
            [0.090, 0.030],
            [0.160, 0.310],
            [0.100, 0.280],
            [0.320, 0.460],
            [0.590, 0.780],
        ], '-^', 'coral'),
        ('Timestamp-based strategy', [
            [0.0, 0.0],
            [0.4, 1.9],
            [4.2, 3.9],
            [3.4, 5.2],
            [5.8, 7.64],
            [6.7, 8.28],
            [7.57, 10.14],
            [12.0, 14.0],
            [20.6, 18.89],
            [16.87, 17.87],
            [16.7, 15.1],
            [21.9, 19.8],
        ], '-s', 'steelblue'),
    ]
    for name, entries, style, color in thru:
        thru = []
        errs = []
        for item in entries:
            thru.append((item[0] + item[1]) / 2.0)
            errs.append(abs(item[0] - item[1]))
        ax.errorbar(replicaNo, thru, yerr=errs, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                    markersize=6)
        ax.set_ylabel("Reordered Commands Ratio (%)")
        ax.legend(loc='best', fancybox=True, frameon=False, framealpha=0.8)
        ax.set_xticks(xticks)
        ax.set_ylim([0, 30])
        ax.set_xticklabels(xticks_label)
        ax.set_xticklabels(("", "", "", "", "", "", "", "", "", "", "", ""))
    ax.set_xticklabels(xticks_label)
    ax.grid(linestyle='--', alpha=0.3)
    f.text(0.5, 0.02, 'Number of Byzantine Nodes', ha='center', va='center')
    plt.savefig('adversarial-comparison.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
