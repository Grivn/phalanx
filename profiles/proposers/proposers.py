import matplotlib.pyplot as plt


def do_plot():
    f, ax = plt.subplots(2, 1, figsize=(6, 6))
    proposer_no = [1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 16, 24, 32]
    x_ticks = [1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 16, 24, 32]
    x_ticks_label = ["1", "2", "3", "4", "5", "6", "7", "8", "10", "12", "16", "24", "32"]

    thru = [
        ('Throughput', [
            (22.5672, 22.3672),
            (39.0164, 39.1564),
            (48.2321, 48.5491),
            (55.2511, 55.7511),
            (62.4261, 62.2607),
            (69.9104, 70.0507),
            (72.2813, 72.8293),
            (78.0622, 78.2642),
            (82.4251, 82.3872),
            (89.4231, 89.6321),
            (101.2412, 101.0927),
            (124.4214, 124.2193),
            (125.2141, 125.9091),
        ], '-o', 'coral'),
    ]
    for name, entries, style, color in thru:
        thru = []
        errs = []
        for item in entries:
            thru.append((item[0] + item[1]) / 2.0)
            errs.append(abs(item[0] - item[1]))
        ax[0].errorbar(proposer_no, thru, yerr=errs, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                       markersize=6)
        ax[0].set_ylabel("throughput ($10^4$ tx/s)")
        ax[0].legend(loc='upper left', fancybox=True, frameon=False, framealpha=0.8)
        ax[0].set_xticks(x_ticks)
        ax[0].set_ylim([0, 150])
        ax[0].set_xticklabels(x_ticks_label)
        ax[0].set_xticklabels(("", "", "", "", "", "", "", "", "", "", "", "", ""))

    cpu_rate = [
        ('Safe Rate', [
            (55.5, 55.5),
            (98.1, 98.1),
            (93.4, 93.4),
            (92.1, 92.1),
            (91.2, 91.2),
            (90.7, 90.2),
            (77.2, 77.6),
            (66.4, 66.2),
            (52.1, 52.7),
            (47.9, 47.1),
            (44.2, 45.2),
            (54.9, 55.9),
            (31.2, 30.2),
        ], '-d', 'darkseagreen'),
    ]
    for name, entries, style, color in cpu_rate:
        safe = []
        errs = []
        for item in entries:
            safe.append((item[0] + item[1]) / 2.0)
            errs.append(abs(item[0] - item[1]))
        ax[1].errorbar(proposer_no, safe, yerr=errs, fmt=style, color=color, mec=color, mfc='none', label='%s' % name,
                       markersize=6)
        ax[1].set_ylabel("safe rate (%)")
        ax[1].legend(loc='upper right', fancybox=True, frameon=False, framealpha=0.8)
        ax[1].set_xticks(proposer_no)
        ax[1].set_xticks(x_ticks)
        ax[1].set_ylim([0, 100])
        ax[1].set_xticklabels(x_ticks_label)
    ax[0].grid(linestyle='--', alpha=0.3)
    ax[1].grid(linestyle='--', alpha=0.3)
    f.text(0.5, 0.04, 'Number of Proposer per Node', ha='center', va='center')
    plt.savefig('proposers.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
