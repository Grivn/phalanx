import matplotlib.pyplot as plt

bsize = [
    ('safe rate', [
        (0, 80.53),
        (1, 67.317767),
        (2, 48.296150),
        (3, 26.824851),
        (4, 16.280314),
        (5, 10.566827),
        (6, 7.110083),
        (7, 8.284638),
        (8, 7.230297),
        (9, 6.139687),
        (10, 5.245343),
    ], '-d', 'peru'),
    ('attacked rate', [
        (0, 0),
        (1, 0.000000),
        (2, 0.000000),
        (3, 0.000000),
        (4, 1.973371),
        (5, 3.240367),
        (6, 9.788574),
        (7, 19.255443),
        (8, 40.349011),
        (9, 61.863760),
        (10, 73.116621),
    ], '-p', 'brown'),
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
    plt.grid(linestyle='--', alpha=0.6, linewidth='1')
    plt.ylim([0, 100])
    plt.ylabel('Rate (%)')
    plt.xlabel('Number of Byzantine Nodes.')
    plt.tight_layout()
    plt.savefig('attacked.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
