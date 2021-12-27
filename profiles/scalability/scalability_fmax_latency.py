import matplotlib.pyplot as plt

bsize = [
    ('Generate-Ordering-Log', [
        (4, 21.1294),
        (5, 24.1294),
        (6, 26.3645),
        (7, 29.3645),
        (8, 33.1018),
        (9, 43.1018),
        (10, 44.7026),
    ], '-o', 'coral'),
    ('Commit-Ordering-Log', [
        (4, 26.1294),
        (5, 36.1294),
        (6, 44.3645),
        (7, 55.3645),
        (8, 69.1018),
        (9, 84.1018),
        (10, 105.7026),
    ], '-o', 'brown'),
    ('Commit-Command-Info', [
        (4, 18.1294),
        (5, 23.1294),
        (6, 26.3645),
        (7, 27.3645),
        (8, 39.1018),
        (9, 59.1018),
        (10, 62.7026),
    ], '-p', 'darkseagreen'),
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
    ax.legend(loc='upper left', fancybox=True, frameon=False, framealpha=0.8)
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0, 120])
    plt.ylabel('Latency (ms)')
    plt.xlabel('Number of Consensus Nodes')
    plt.tight_layout()
    plt.savefig('scalability_fmax_latency.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
