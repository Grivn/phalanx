import matplotlib.pyplot as plt


def do_plot():
    f, ax = plt.subplots(2, 1, figsize=(6, 6))
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
        ax[0].errorbar(replica_no, thru, yerr=0, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                       markersize=6)
        ax[0].set_ylabel("throughput ($10^3$ tx/s)")
        ax[0].legend(loc='lower center', fancybox=True, frameon=False, framealpha=0.8)
        ax[0].set_xticks(x_ticks)
        ax[0].set_ylim([0, 40])
        ax[0].set_xticklabels(x_ticks_label)
        ax[0].set_xticklabels(("", "", "", ""))

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
        ax_real = ax[0].twinx()
        ax_real.errorbar(replica_no, real, yerr=0, fmt=style, mec=color, color=color, mfc='none', label='%s' % name,
                         markersize=6)
        ax_real.set_ylabel("blocks per HS-commit")
        ax_real.legend(loc='lower right', fancybox=True, frameon=False, framealpha=0.8)
        ax_real.set_xticks(x_ticks)
        ax_real.set_ylim([10, 30])
        ax_real.set_xticklabels(x_ticks_label)
        ax_real.set_xticklabels(("", "", "", ""))

    cpu_rate = [
        ('Safe Rate', [
            (81.120102, 55.5),
            (71.447033, 98.1),
            (63.439294, 93.4),
            (58.203584, 92.1),
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
    f.text(0.5, 0.04, 'Number of Proposers', ha='center', va='center')
    plt.savefig('proposers_new.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
