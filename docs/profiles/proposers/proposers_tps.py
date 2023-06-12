import matplotlib.pyplot as plt


def do_plot():
    f, ax = plt.subplots(1, 1, figsize=(6, 3))
    replica_no = [4, 8, 12, 16]
    x_ticks = [4, 8, 12, 16]
    x_ticks_label = ["4", "8", "12", "16"]

    thru = [
        ('Throughput', [
            (23.8784, 548784.417931),
            (26.2163, 532163.059013),
            (28.1742, 501742.085271),
            (29.5932, 555932.468898),
        ], '-d', 'coral'),

    ]
    for name, entries, style, color in thru:
        thru = []
        for item in entries:
            thru.append(item[0])
        ax.errorbar(replica_no, thru, yerr=0, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                    markersize=6)
        ax.set_ylabel("throughput ($10^3$ tx/s)")
        ax.legend(loc='lower center', fancybox=True, frameon=False, framealpha=0.8)
        ax.set_xticks(x_ticks)
        ax.set_ylim([0, 40])

    real_block = [
        ('Block Rate', [
            18.7773592195498888,
            21.150943396226415,
            23.285630153121319,
            26.634081551334839,
        ], '-o', 'burlywood'),
    ]
    for name, entries, style, color in real_block:
        real = []
        for item in entries:
            real.append(item)
        ax_real = ax.twinx()
        ax_real.errorbar(replica_no, real, yerr=0, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                         markersize=6)
        ax_real.set_ylabel("blocks per HS-commit")
        ax_real.legend(loc='lower right', fancybox=True, frameon=False, framealpha=0.8)
        ax_real.set_xticks(x_ticks)
        ax_real.set_ylim([10, 30])
        ax_real.set_xticklabels(x_ticks_label)
        ax_real.set_xticklabels(x_ticks_label)
    ax.grid(linestyle='--', alpha=0.3)
    f.text(0.5, 0.02, 'Number of Proposers', ha='center', va='center')
    plt.savefig('proposers_tps.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
