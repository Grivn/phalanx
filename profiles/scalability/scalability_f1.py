import matplotlib.pyplot as plt


def do_plot():
    f, ax = plt.subplots(2, 1, figsize=(6, 6))
    replica_no = [4, 5, 6, 7, 8, 9, 10]
    x_ticks = [4, 5, 6, 7, 8, 9, 10]
    x_ticks_label = ["4", "5", "6", "7", "8", "9", "10"]

    thru = [
        ('Throughput', [
            (113.8784, 548784.417931),
            (112.2163, 532163.059013),
            (115.1742, 501742.085271),
            (119.5932, 555932.468898),
            (104.0121, 550121.048102),
            (92.0143, 590143.754558),
            (83.6520, 606520.018519),
        ], '-d', 'coral'),

    ]
    for name, entries, style, color in thru:
        thru = []
        for item in entries:
            thru.append(item[0])
        ax[0].errorbar(replica_no, thru, yerr=0, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                       markersize=6)
        ax[0].set_ylabel("throughput ($10^4$ tx/s)")
        ax[0].legend(loc='lower center', fancybox=True, frameon=False, framealpha=0.8)
        ax[0].set_xticks(x_ticks)
        ax[0].set_ylim([0, 130])
        ax[0].set_xticklabels(x_ticks_label)
        ax[0].set_xticklabels(("", "", "", "", "", "", ""))

    real_block = [
        ('Block Rate', [
            5.7773592195498888,
            8.150943396226415,
            10.285630153121319,
            12.634081551334839,
            14.921107472462042,
            16.5292141396435874,
            17.8232738557020944,
        ], '-o', 'burlywood'),
    ]
    for name, entries, style, color in real_block:
        real = []
        for item in entries:
            real.append(item)
        ax_real = ax[0].twinx()
        ax_real.errorbar(replica_no, real, yerr=0, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                         markersize=6)
        ax_real.set_ylabel("blocks per HS-commit")
        ax_real.legend(loc='lower right', fancybox=True, frameon=False, framealpha=0.8)
        ax_real.set_xticks(x_ticks)
        ax_real.set_ylim([0, 20])
        ax_real.set_xticklabels(x_ticks_label)
        ax_real.set_xticklabels(("", "", "", "", "", "", ""))

    cpu_rate = [
        ('Attacked Rate', [
            (4.120102, 55.5),
            (1.947033, 98.1),
            (0.639294, 93.4),
            (0.803584, 92.1),
            (0.552313, 91.2),
            (0.314721, 90.2),
            (0.421610, 77.6),
        ], '-d', 'darkseagreen'),
    ]
    for name, entries, style, color in cpu_rate:
        cpus = []
        for item in entries:
            cpus.append(item[0])
        ax[1].errorbar(replica_no, cpus, yerr=0, fmt=style, color=color, mec=color, mfc='none', label='%s' % name,
                       markersize=6)
        ax[1].set_ylabel("attacked rate (%)")
        ax[1].legend(loc='upper right', fancybox=True, frameon=False, framealpha=0.8)
        ax[1].set_xticks(replica_no)
        ax[1].set_xticks(x_ticks)
        ax[1].set_ylim([0, 10])
        ax[1].set_xticklabels(x_ticks_label)
    ax[0].grid(linestyle='--', alpha=0.3)
    ax[1].grid(linestyle='--', alpha=0.3)
    f.text(0.5, 0.04, 'Number of Consensus Nodes.', ha='center', va='center')
    plt.savefig('scalability_f1.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
