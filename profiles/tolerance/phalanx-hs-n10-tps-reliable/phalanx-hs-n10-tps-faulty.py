import matplotlib.pyplot as plt

bsize = [
    ('Safe-Rate', [
        (25, 77.133234),
        (50, 43.721234),
        (75, 23.109017),
        (93, 10.527314),
    ], '-o', 'coral'),
]


def do_plot():
    f = plt.figure(1, figsize=(6, 3))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in bsize:
        throughput = []
        replica_no = []
        for c, t in entries:
            throughput.append(t)
            replica_no.append(c)
        ax.plot(replica_no, throughput, style, color=color, label='%s' % name, markersize=6, alpha=0.8)
    plt.legend(loc='upper right', fancybox=True, frameon=False, framealpha=0.8)
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0, 100])
    plt.ylabel('Rate (%)')
    plt.xlabel('Throughput ($10^4$tx/s)')
    plt.tight_layout()
    plt.savefig('front-attack-rate-tps.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
