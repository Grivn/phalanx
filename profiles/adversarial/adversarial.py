import matplotlib.pyplot as plt

bsize = [
    ('phalanx reordered rate', [
        (0, 0), # 0
        (1, 0), # 0
        (2, 0), # 0
        (3, 0), # 0
        (4, 0), # 0.028
        (5, 0), # 0.035
        (6, 0), # 0.04
        (7, 0.09), # 0.03
        (8, 0.16), # 0.31
        (9, 0.1), # 0.28
        (10, 0.32), #  0.46
        (11, 0.59), # 0.78
    ], '-p', 'brown'),
    ('timestamp-based reordered rate', [
        (0, 0), # 0
        (1, 0.4), # 1.9
        (2, 4.2), # 3.9
        (3, 3.4), # 5.2
        (4, 4.8), # 7.64
        (5, 5.7), # 8.28
        (6, 5.57), # 10.14
        (7, 12), # 14
        (8, 20.6), # 18.89
        (9, 16.87), # 17.87
        (10, 16.7), # 15.1
        (11, 21.9), # 19.0
    ], '-d', 'peru'),
]


def do_plot():
    f = plt.figure(1, figsize=(6, 4))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in bsize:
        throughput = []
        replica_no = []
        for c, t in entries:
            throughput.append(t)
            replica_no.append(c)
        ax.plot(replica_no, throughput, style, color=color, label='%s' % name, markersize=6, alpha=0.8)
    plt.legend(loc='upper left', fancybox=True, frameon=False, framealpha=0.8)
    plt.grid(linestyle='--', alpha=0.6, linewidth='1')
    plt.ylim([0, 100])
    plt.ylabel('Rate (%)')
    plt.xlabel('Number of Byzantine Nodes.')
    plt.tight_layout()
    plt.savefig('adversarial.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
