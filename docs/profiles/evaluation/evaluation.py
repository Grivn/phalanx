import matplotlib.pyplot as plt

bsize = [
    ('HS1000', [
        (4, 21.2944),
        (5, 17.3973),
        (6, 16.6320),
        (7, 15.5899),
        (8, 15.5608),
        (9, 14.9115),
        (10, 14.9119),
        (11, 14.4893),
        (12, 14.0322),
        (13, 14.1527),
        (14, 13.6612),
        (15, 13.0852),
    ], '-o', 'coral'),
    ('HS4000', [
        (4, 32.1294),
        (5, 27.5342),
        (6, 26.3645),
        (7, 26.1869),
        (8, 27.1018),
        (9, 26.2143),
        (10, 25.7026),
        (11, 25.4374),
        (12, 24.6564),
        (13, 25.6830),
        (14, 24.5260),
        (15, 23.3333),
    ], '-o', 'coral'),
    ('Phalanx-HS1000', [
        (4, 54.8784),
        (5, 53.2163),
        (6, 50.1742),
        (7, 55.5932),
        (8, 55.0121),
        (9, 59.0143),
        (10, 60.6520),
        (11, 59.1436),
        (12, 56.3430),
        (13, 59.6944),
        (14, 56.4608),
        (15, 54.0189),
    ], '-p', 'darkseagreen')]


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
    plt.legend(fancybox=True, frameon=False, framealpha=0.8, mode={"expand", None}, ncol=3, loc='upper center')
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0, 70])
    plt.ylabel('Throughput ($10^4$tx/s)')
    plt.xlabel('Replica Number')
    plt.tight_layout()
    plt.savefig('evaluation.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
