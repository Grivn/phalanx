import matplotlib.pyplot as plt


def do_plot():
    f, ax = plt.subplots(2, 1, figsize=(6, 6))
    replica_no = [4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]
    x_ticks = [4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]
    x_ticks_label = ["4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"]

    thru = [
        ('Throughput', [
            (54.8784, 548784.417931),
            (53.2163, 532163.059013),
            (50.1742, 501742.085271),
            (55.5932, 555932.468898),
            (55.0121, 550121.048102),
            (59.0143, 590143.754558),
            (60.6520, 606520.018519),
            (59.1436, 591436.034703),
            (56.3430, 563430.543858),
            (59.6944, 596944.303082),
            (56.4608, 564608.555150),
            (54.0189, 540189.9915136739),
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
        ax[0].set_ylim([0, 65])
        ax[0].set_xticklabels(x_ticks_label)
        ax[0].set_xticklabels(("", "", "", "", "", "", "", "", "", "", "", ""))

    real_block = [
        ('Block Rate', [
            1.7773592195498888,
            2.150943396226415,
            2.285630153121319,
            2.634081551334839,
            2.921107472462042,
            3.5292141396435874,
            3.8232738557020944,
            4.187312186978297,
            4.427184466019417,
            4.787277448071217,
            5.0900932918702795,
            5.635924369747899,
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
        ax_real.set_ylim([0, 6])
        ax_real.set_xticklabels(x_ticks_label)
        ax_real.set_xticklabels(("", "", "", "", "", "", "", "", "", "", "", ""))

    cpu_rate = [
        ('Safe Rate', [
            (90.120102, 55.5),
            (98.447033, 98.1),
            (99.439294, 93.4),
            (76.203584, 92.1),
            (89.152313, 91.2),
            (98.714721, 90.2),
            (96.221610, 77.6),
            (97.406253, 66.2),
            (98.033933, 52.7),
            (95.237310, 47.1),
            (95.302763, 45.2),
            (95.251208, 55.9),
        ], '-d', 'darkseagreen'),
    ]
    for name, entries, style, color in cpu_rate:
        cpus = []
        for item in entries:
            cpus.append(item[0])
        ax[1].errorbar(replica_no, cpus, yerr=0, fmt=style, color=color, mec=color, mfc='none', label='%s' % name,
                       markersize=6)
        ax[1].set_ylabel("safe rate (%)")
        ax[1].legend(loc='lower right', fancybox=True, frameon=False, framealpha=0.8)
        ax[1].set_xticks(replica_no)
        ax[1].set_xticks(x_ticks)
        ax[1].set_ylim([0, 100])
        ax[1].set_xticklabels(x_ticks_label)
    ax[0].grid(linestyle='--', alpha=0.3)
    ax[1].grid(linestyle='--', alpha=0.3)
    f.text(0.5, 0.04, 'Number of Node', ha='center', va='center')
    plt.savefig('scalability.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
